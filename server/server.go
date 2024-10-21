package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	channelz "github.com/rantav/go-grpc-channelz"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	channelzsvc "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"

	"github.com/tschroed/spellingbee"
	pb "github.com/tschroed/spellingbee/server/proto"
)

const (
	DEBUG  = false
	otRoot = "zweknu.org/spellingbee"
)

var (
	pFlag    = flag.Int("p", 3000, "gRPC port")
	wFlag    = flag.Int("w", 3001, "Web server port")
	tFlag    = flag.String("t", "page_html.tmpl", "Page template")
	meter    = otel.Meter(otRoot)
	logger   = otelslog.NewLogger(otRoot)
	solveCnt metric.Int64Counter
)

func debug(v any) {
	if DEBUG {
		log.Println(v)
	}
}

func readWords(fname string) ([]string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	words := make([]string, 0)
	for l, _, err := r.ReadLine(); err != io.EOF; l, _, err = r.ReadLine() {
		words = append(words, strings.ToLower(strings.TrimSpace(string(l))))
	}
	slices.Sort(words)
	words = slices.Compact(words)
	return words, nil
}

func readTemplate(fname string) (string, error) {
	b, err := os.ReadFile(fname)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func usage() {
	log.Fatalf("usage: %s [-p <port>] <dictionary>\n", os.Args[0])
}

type server struct {
	pb.UnimplementedSpellingbeeServer
	dict *spellingbee.Dictionary
}

func (s *server) FindWords(ctx context.Context, in *pb.SpellingbeeRequest) (*pb.SpellingbeeReply, error) {
	solveValueAttr := attribute.String("solve.mode", "grpc")
	solveCnt.Add(ctx, 1, metric.WithAttributes(solveValueAttr))
	soln := s.dict.FindWords(ctx, in.Letters)
	slices.SortFunc(soln, spellingbee.CmpFn(in.Letters, in.Reverse))
	return &pb.SpellingbeeReply{Words: soln}, nil
}

func mtime(fname string) (time.Time, error) {
	st, err := os.Stat(fname)
	if err != nil {
		return time.Unix(0, 0), err
	}
	return st.ModTime(), nil
}

type webApp struct {
	tmpl *template.Template
	dict *spellingbee.Dictionary
}

func (a *webApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Letters string
		Reverse bool
		Soln    []string
	}
	data.Letters = r.FormValue("letters")
	if v := r.FormValue("reverse"); v != "" {
		if b, err := strconv.ParseBool(v); err != nil {
			log.Println(err)
		} else {
			data.Reverse = b
		}
	}
	if data.Letters != "" {
		solveValueAttr := attribute.String("solve.mode", "web")
		solveCnt.Add(r.Context(), 1, metric.WithAttributes(solveValueAttr))
		soln := a.dict.FindWords(r.Context(), data.Letters)
		slices.SortFunc(soln, spellingbee.CmpFn(data.Letters, data.Reverse))
		data.Soln = soln
	}
	if a.tmpl != nil {
		a.tmpl.Execute(w, data)
	}
}

func init() {
	var err error
	solveCnt, err = meter.Int64Counter("spellingbee.solves",
		metric.WithDescription("Number of calls to solve a puzzle"),
		metric.WithUnit("{solve}"))
	if err != nil {
		panic(err)
	}
}

type dictStats struct {
	size    metric.Int64Gauge
	solnCnt metric.Int64Counter
}

func (s *dictStats) RecordSize(ctx context.Context, size int) {
	s.size.Record(ctx, int64(size))
}
func (s *dictStats) RecordSolutions(ctx context.Context, soln []string) {
	solnLenAttr := attribute.Int("solution.len", len(soln))
	s.solnCnt.Add(ctx, 1, metric.WithAttributes(solnLenAttr))
}
func newDictStats() *dictStats {
	size, err := meter.Int64Gauge("spellingbee.dict_size",
		metric.WithDescription("Size of spellingbee dictionary"))
	if err != nil {
		panic(err)
	}
	solnCnt, err := meter.Int64Counter("spellingbee.solutions",
		metric.WithDescription("Number of solutions by solution set size"),
		metric.WithUnit("{solution}"))
	if err != nil {
		panic(err)
	}
	return &dictStats{
		size:    size,
		solnCnt: solnCnt,
	}

}

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	/* TODO(trevors): Figure out why ^C doesn't seem to be invoking signal
	* handling as expected.
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer func() {
			log.Printf("Interrupt received, quitting...")
			stop()
			os.Exit(0)
		}()
	*/
	ctx := context.Background()

	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		usage()
	}
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()
	words, err := readWords(args[0])
	if err != nil {
		log.Fatalf("%v", err)
	}
	dict := spellingbee.NewDictionary(ctx, words, newDictStats())
	debug(dict)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *pFlag))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSpellingbeeServer(s, &server{dict: dict})
	reflection.Register(s)
	channelzsvc.RegisterChannelzServiceToServer(s)

	// Set up web app
	var tmpl *template.Template
	if t, err := readTemplate(*tFlag); err != nil {
		log.Println(err)
	} else {
		log.Println("Parsing template...")
		tmpl, err = template.New("page").Parse(t)
		if err != nil {
			panic(err)
		}
	}
	http.Handle("/", &webApp{tmpl: tmpl, dict: dict})
	if err != nil {
		log.Fatal(err)
	}
	// Set up a channelz ui at /debug/channelz/
	a := lis.Addr()
	http.Handle("/debug/", channelz.CreateHandler("/debug", a.String()))
	// Listen on wFlag
	wlis, err := net.Listen("tcp", fmt.Sprintf(":%d", *wFlag))
	if err != nil {
		log.Fatal(err)
	}
	go http.Serve(wlis, nil)

	mt, err := mtime(os.Args[0])
	if err != nil {
		log.Printf("unable to get mtime of %s: %v", os.Args[0], err)
	}
	log.Printf("Server (mtime %v) listening at %v", mt, a)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

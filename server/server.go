package main

import (
	"bufio"
	"context"
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

	"google.golang.org/grpc"
	channelzsvc "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
	"github.com/rantav/go-grpc-channelz"


	"github.com/tschroed/spellingbee"
	pb "github.com/tschroed/spellingbee/server/proto"
)

const (
	DEBUG = false
)

var (
	pFlag = flag.Int("p", 3000, "gRPC port")
	wFlag = flag.Int("w", 3001, "Web server port")
	tFlag = flag.String("t", "page_html.tmpl", "Page template")
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
		words = append(words, strings.ToLower(string(l)))
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

func (s *server) FindWords(_ context.Context, in *pb.SpellingbeeRequest) (*pb.SpellingbeeReply, error) {
	soln := s.dict.FindWords(in.Letters)
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
		Soln []string
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
		soln := a.dict.FindWords(data.Letters)
		slices.SortFunc(soln, spellingbee.CmpFn(data.Letters, data.Reverse))
		data.Soln = soln
	}
	if a.tmpl != nil {
		a.tmpl.Execute(w, data)
	}
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		usage()
	}
	words, err := readWords(args[0])
	if err != nil {
		log.Fatalf("%v", err)
	}
	dict := spellingbee.NewDictionary(words)
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
	http.Handle("/",  &webApp{tmpl: tmpl, dict: dict})
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

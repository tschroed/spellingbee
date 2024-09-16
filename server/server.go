package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"slices"
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
	aFlag = flag.Int("a", 3001, "Admin port")
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
	a := lis.Addr()

	// Setup a channelz ui at /debug/channelz/ listening on port aFlag.
	http.Handle("/", channelz.CreateHandler("/debug", a.String()))
	alis, err := net.Listen("tcp", fmt.Sprintf(":%d", *aFlag))
	if err != nil {
		    log.Fatal(err)
	}
	go http.Serve(alis, nil)

	mt, err := mtime(os.Args[0])
	if err != nil {
		log.Printf("unable to get mtime of %s: %v", os.Args[0], err)
	}
	log.Printf("Server (mtime %v) listening at %v", mt, a)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"slices"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/tschroed/spellingbee"
	pb "github.com/tschroed/spellingbee/server/proto"
)

const (
	DEBUG = false
)

var pFlag = flag.Int("p", 3000, "Port to listen on")

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
	dict spellingbee.Dictionary
}

func (s *server) FindWords(_ context.Context, in *pb.SpellingbeeRequest) (*pb.SpellingbeeReply, error) {
	soln := spellingbee.FindWords(s.dict, in.Letters)
	sortFn := func(reverse bool) func(string, string) int {
		reta := 1
		retb := -1
		if reverse {
			reta, retb = retb, reta
		}
		return func(a, b string) int {
			la := len(a)
			lb := len(b)
			if la < lb {
				return reta
			}
			if lb < la {
				return retb
			}
			return 0
		}
	}
	slices.SortFunc(soln, sortFn(in.Reverse))
	return &pb.SpellingbeeReply{Words: soln}, nil
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
	d := spellingbee.BuildDictionary(words)
	debug(d)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *pFlag))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSpellingbeeServer(s, &server{dict: d})
	reflection.Register(s)
	a := lis.Addr()
	log.Printf("Server listening at %v", a)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

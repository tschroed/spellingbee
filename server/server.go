package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/tschroed/spellingbee"
)

const (
	DEBUG = false
)

var pFlag = flag.Int("p", 3000, "Port to listen on")

func debug(v any) {
	if DEBUG {
		fmt.Println(v)
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
	fmt.Printf("usage: %s [-p <port>] <dictionary>\n", os.Args[0])
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		usage()
		os.Exit(1)
	}
	words, err := readWords(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// debug(charBits)
	wordKeys := make(map[string]spellingbee.Key, 0)
	for _, word := range words {
		k := spellingbee.KeyOf(word)
		if k == 0 {
			continue
		}
		wordKeys[word] = k
	}
	debug(wordKeys)
	soln := spellingbee.FindWords(wordKeys, spellingbee.KeyOf(args[1]))
	fmt.Println(soln)
}

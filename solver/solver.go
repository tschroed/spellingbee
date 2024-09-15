package main

import (
	"bufio"
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
	fmt.Printf("usage: %s <dictionary> <letters>\n", os.Args[0])
}

func main() {
	if len(os.Args) != 3 {
		usage()
		os.Exit(1)
	}
	words, err := readWords(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// debug(charBits)
	wordsKeys := spellingbee.BuildDictionary(words)
	debug(wordsKeys)
	soln := spellingbee.FindWords(wordsKeys, os.Args[2])
	fmt.Println(soln)
}

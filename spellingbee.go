package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/exp/slices"
)

const (
	keyLen = 7
	DEBUG  = false
)

var charBits = map[string]uint32{
	"a": 1 << 0,
	"b": 1 << 1,
	"c": 1 << 2,
	"d": 1 << 3,
	"e": 1 << 4,
	"f": 1 << 5,
	"g": 1 << 6,
	"h": 1 << 7,
	"i": 1 << 8,
	"j": 1 << 9,
	"k": 1 << 10,
	"l": 1 << 11,
	"m": 1 << 12,
	"n": 1 << 13,
	"o": 1 << 14,
	"p": 1 << 15,
	"q": 1 << 16,
	"r": 1 << 17,
	"s": 1 << 18,
	"t": 1 << 19,
	"u": 1 << 20,
	"v": 1 << 21,
	"w": 1 << 22,
	"x": 1 << 23,
	"y": 1 << 24,
	"z": 1 << 25,
}

func debug(v any) {
	if DEBUG {
		fmt.Println(v)
	}
}

func usage() {
	fmt.Printf("usage: %s <dictionary>\n", os.Args[0])
}

type key uint32

func keyOf(s string) key {
	debug(s)
	c := strings.Split(strings.ToLower(s), "")
	slices.Sort(c)
	c = slices.Compact(c)
	var k uint32
	for _, ch := range c {
		v, ok := charBits[ch]
		if !ok {
			return 0
		}
		k |= v
	}
	debug(k)
	return key(k)
}

func readWords(fname string) ([]string, error) {
	f, err := os.Open(os.Args[1])
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

// If haystack key contains the needle key, return true.
// The first letter of haystack is the central letter and must be present.
func keyContains(haystack, needle key) bool {
	if haystack == 0 || needle == 0 {
		return false
	}
	return (haystack & needle) == needle
}

func findWords(words map[string]key, k key) []string {
	soln := make([]string, 0)
	for w, wk := range words {
		if keyContains(k, wk) {
			soln = append(soln, w)
		}
	}
	return soln
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
	debug(charBits)
	wordKeys := make(map[string]key, 0)
	for _, word := range words {
		wordKeys[word] = keyOf(word)
	}
	soln := findWords(wordKeys, keyOf(os.Args[2]))
	fmt.Println(soln)
}

/* Related but unnecessary.

type set map[string]struct{}

func findKey(keys set, word string) (string, error) {
        wk := keyOf(word)
        if len(wk) == keyLen {
                return wk, nil
        }
        for k, _ := range(keys) {
                if keyContains(k, wk) {
                        return k, nil
                }
        }
        return "", fmt.Errorf("no key for %s", word)
}

func buildKeys(words []string) set {
        keys := make([]string, 0)
        for _, w := range words {
                k := keyOf(w)
                if len(k) == keyLen {
                        keys = append(keys, k)
                }
        }
        s := make(map[string]struct{}, 0)
        for _, k := range keys {
                s[k] = struct{}{}
        }
        return s
}

*/

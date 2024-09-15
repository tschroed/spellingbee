package spellingbee

// TODO: write some tests!

import (
	"fmt"
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

type key uint32
type Dictionary map[string]key

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

func BuildDictionary(words []string) Dictionary {
	wk := make(Dictionary, 0)
	for _, word := range words {
		k := keyOf(word)
		if k == 0 {
			continue
		}
		wk[word] = k
	}
	return wk
}

// If haystack key contains the needle key, return true.
// The first letter of haystack is the central letter and must be present.
func keyContains(haystack, needle key) bool {
	if haystack == 0 || needle == 0 {
		return false
	}
	return (haystack & needle) == needle
}

func FindWords(wordsKeys Dictionary, letters string) []string {
	lk := keyOf(letters)
	soln := make([]string, 0)
	for w, wk := range wordsKeys {
		if w[0] == letters[0] && keyContains(lk, wk) {
			soln = append(soln, w)
		}
	}
	return soln
}

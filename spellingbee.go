package spellingbee

import (
	"log"
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
		log.Println(v)
	}
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

// If haystack key contains the needle key, return true.
// The first letter of haystack is the central letter and must be present.
func keyContains(haystack, needle key) bool {
	if haystack == 0 || needle == 0 {
		return false
	}
	return (haystack & needle) == needle
}

// Get a comparison function suitable for SortFunc based on pangrams and length.
func CmpFn(letters string, reverse bool) func(a, b string) int {
	kl := keyOf(letters)
	reta := -1
	retb := 1
	if reverse {
		reta, retb = retb, reta
	}
	return func(a, b string) int {
		ka := keyOf(a)
		la := len(a)
		kb := keyOf(b)
		lb := len(b)

		// If only one is a pangram, put it at the end
		if ka == kl && kb != kl {
			return retb
		}
		if kb == kl && ka != kl {
			return reta
		}
		// Otherwise, neither or both are pangrams so sort based on length.
		if la < lb {
			return reta
		}
		if la > lb {
			return retb
		}
		return 0
	}
}

// If the keys are sparse this is memory-inefficience because there are a
// lot of []strings but searching is generally faster and multiple hits can
// be found simultaneously. For example, [moisturize moisturizer
// moisturizers moisturizes] are a single equivalence class. In practice it
// has resulted in a 25% to 50% speed-up in searches.
type keyswords map[key][]string
type Dictionary struct {
	kw keyswords
}

func NewDictionary(words []string) *Dictionary {
	kw := make(map[key][]string, 0)
	for _, word := range words {
		k := keyOf(word)
		if k == 0 {
			continue
		}
		arr, ok := kw[k]
		if !ok {
			arr = make([]string, 0)
		}
		arr = append(arr, word)
		kw[k] = arr
	}
	debug(kw)
	debug(len(kw))
	return &Dictionary{kw: kw}
}

func (d *Dictionary) FindWords(letters string) []string {
	lk := keyOf(letters)
	rk := keyOf(string(letters[0])) // Must be in every returned word.
	soln := make([]string, 0)
	for kw, w := range d.kw {
		if keyContains(kw, rk) && keyContains(lk, kw) {
			soln = append(soln, w...)
		}
	}
	return soln
}

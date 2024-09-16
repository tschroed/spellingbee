package spellingbee

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
)

/* Tests of the exported API. These should be relatively stable. API
 * coverage should be 100%. */

func TestNewDictionary(t *testing.T) {
	cases := []struct{
		name string
		words []string
		want wordskeys
	}{
		{
			name: "unique keys",
			words: []string{"foo", "bar", "baz"},
			want: wordskeys{"bar": 131075, "baz": 33554435, "foo": 16416},
		},
		{
			name: "equivalent keys",
			words: []string{"foobara", "foobar", "foobaroo"},
			want: wordskeys{"foobara": 147491, "foobar": 147491, "foobaroo":   147491},
		},
	}

	for _, tc := range cases {
		got := NewDictionary(tc.words)
		if diff := cmp.Diff(tc.want, got.wk); diff != "" {
			t.Errorf("%s: NewDictionary(%v) mismatch (-want +got):\n%s", tc.name, tc.words, diff)
		}
	}
}

func TestFindWords(t *testing.T) {
	letters := "alphynx"
	// Note that "ply" does not contain the mandatory "a".
	words := []string{"alpha", "beta", "gamma", "ply", "phalanx", "philistine", "alfalfa", "pharynx"}
	d := NewDictionary(words)
	got := d.FindWords(letters)
	want := []string{"alpha", "phalanx"}
	slices.Sort(got)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FindWords() mismatch (-want +got):\n%s", diff)
	}
}

/* Tests of internal implementations. Not a stable part of the API but can
 * be useful to test anyhow in order to assure that components of the
 * implementation are working as expected. */

 func TestKeyOf(t *testing.T) {
	 k1 := keyOf("foo")
	 k2 := keyOf("bar")
	 if k1 == k2 {
		 t.Errorf("keyOf(\"foo\") == keyOf(\"bar\")?")
	 }
	 k2 = keyOf("ofo")
	 if k1 != k2 {
		 t.Errorf("keyOf(\"foo\") == keyOf(\"ofo\")?")
	 }
 }

 func TestKeyContains(t *testing.T) {
	 k1 := keyOf("foobar")
	 k2 := keyOf("foo")
	 if !keyContains(k1, k2) {
		 t.Errorf("[true1]: keyContains(%v, %v) == false", k1, k2)
	 }
	 k2 = keyOf("bar")
	 if !keyContains(k1, k2) {
		 t.Errorf("[true2]: keyContains(%v, %v) == false", k1, k2)
	 }
	 k2 = keyOf("baz")
	 if keyContains(k1, k2) {
		 t.Errorf("[false]: keyContains(%v, %v) == true", k1, k2)
	 }
 }

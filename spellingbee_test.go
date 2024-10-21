package spellingbee

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
)

type testStats struct {
	Size      int
	Solutions map[int]int
}

func (s *testStats) RecordSize(_ context.Context, size int) {
	s.Size = size
}
func (s *testStats) RecordSolutions(_ context.Context, soln []string) {
	s.Solutions[len(soln)] += 1
}
func newTestStats() *testStats {
	return &testStats{Solutions: make(map[int]int, 0)}
}

/* Tests of the exported API. These should be relatively stable. API
 * coverage should be 100%. */

func TestNewDictionary(t *testing.T) {
	cases := []struct {
		name  string
		words []string
		want  keyswords
	}{
		{
			name:  "unique keys",
			words: []string{"foo", "bar", "baz"},
			want:  keyswords{131075: {"bar"}, 33554435: {"baz"}, 16416: {"foo"}},
		},
		{
			name:  "equivalent keys",
			words: []string{"foobara", "foobar", "foobaroo"},
			want:  keyswords{147491: {"foobara", "foobar", "foobaroo"}},
		},
	}

	for _, tc := range cases {
		got := NewDictionary(context.Background(), tc.words, nil)
		if diff := cmp.Diff(tc.want, got.kw); diff != "" {
			t.Errorf("%s: NewDictionary(%v) mismatch (-want +got):\n%s", tc.name, tc.words, diff)
		}
	}
}

func TestFindWordsNonEmpty(t *testing.T) {
	// Note that "ply" does not contain the mandatory "a".
	words := []string{"alpha", "beta", "gamma", "ply", "phalanx", "philistine", "alfalfa", "pharynx"}
	d := NewDictionary(context.Background(), words, nil)
	cases := []struct {
		letters string
		want    []string
	}{
		{
			letters: "alphynx",
			want:    []string{"alpha", "phalanx"},
		},
		{
			letters: "fla",
			want:    []string{"alfalfa"},
		},
	}
	for i, tc := range cases {
		got := d.FindWords(context.Background(), tc.letters)
		slices.Sort(got)
		if diff := cmp.Diff(tc.want, got); diff != "" {
			t.Errorf("[%d, %s] FindWords() mismatch (-want +got):\n%s", i, tc.letters, diff)
		}
	}
}

func TestFindWordsEmpty(t *testing.T) {
	words := []string{"alpha", "beta", "gamma", "ply", "phalanx", "philistine", "alfalfa", "pharynx"}
	d := NewDictionary(context.Background(), words, nil)
	got := d.FindWords(context.Background(), "")
	want := []string{}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FindWords() mismatch (-want +got):\n%s", diff)
	}
}

func TestStats(t *testing.T) {
	// Note that "ply" does not contain the mandatory "a".
	words := []string{"alpha", "beta", "gamma", "ply", "phalanx", "philistine", "alfalfa", "pharynx"}
	s := newTestStats()
	d := NewDictionary(context.Background(), words, s)
	if s.Size != 8 {
		t.Errorf("Dictionary size misrecorded, want 8 got %v", s.Size)
	}
	if len(s.Solutions) != 0 {
		t.Errorf("Solutions misrecorded, want {} got %v", s.Solutions)
	}
	cases := []struct {
		letters   string
		wantSoln  []string
		wantStats map[int]int // Cumulative w/ previous test cases
	}{
		{
			letters:   "alphynx",
			wantSoln:  []string{"alpha", "phalanx"},
			wantStats: map[int]int{2: 1},
		},
		{
			letters:   "agfml",
			wantSoln:  []string{"alfalfa", "gamma"},
			wantStats: map[int]int{2: 2},
		},
		{
			letters:   "plha",
			wantSoln:  []string{"alpha"},
			wantStats: map[int]int{1: 1, 2: 2},
		},
	}
	for i, tc := range cases {
		got := d.FindWords(context.Background(), tc.letters)
		slices.Sort(got)
		if diff := cmp.Diff(tc.wantSoln, got); diff != "" {
			t.Errorf("[%d: %s] FindWords() mismatch (-want +got):\n%s", i, tc.letters, diff)
		}
		if diff := cmp.Diff(s.Solutions, tc.wantStats); diff != "" {
			t.Errorf("[%d: %s] stats solutions mismatch (-want +got):\n%s", i, tc.letters, diff)
		}
	}
}

func TestCmpFn(t *testing.T) {
	cases := []struct {
		name   string
		first  string
		second string
		equiv  bool
	}{
		{
			name:   "equivalents",
			first:  "lime",
			second: "lime",
			equiv:  true,
		},
		{
			name:   "length",
			first:  "lime",
			second: "mingling",
		},
		{
			name:   "pangram",
			first:  "lime",
			second: "melding",
		},
		{
			name:   "pangram trumps longer",
			first:  "mingling",
			second: "melding",
		},
		{
			name:   "dual-pangram same length",
			first:  "melding",
			second: "mingled",
			equiv:  true,
		},
		{
			name:   "dual-pangram uses length",
			first:  "melding",
			second: "meddling",
		},
		{
			name:   "higher score can trump pangram",
			first:  "melding",
			second: "limelimelimelime",
		},
	}
	letters := "mdegiln"
	cmp := CmpFn(letters, false)
	cmpr := CmpFn(letters, true)
	for _, tc := range cases {
		v := cmp(tc.first, tc.second)
		vr := cmpr(tc.first, tc.second)
		if tc.equiv {
			if v != 0 {
				t.Errorf("%s: cmp(\"%s\", \"%s\") want 0 got %d", tc.name, tc.first, tc.second, v)
			}
			if vr != 0 {
				t.Errorf("%s: cmpr(\"%s\", \"%s\") want 0 got %d", tc.name, tc.first, tc.second, vr)
			}
			continue
		}
		if v != -1 {
			t.Errorf("%s: cmp(\"%s\", \"%s\") want -1 got %d", tc.name, tc.first, tc.second, v)
		}
		if vr != 1 {
			t.Errorf("%s: cmpr(\"%s\", \"%s\") want 1 got %d", tc.name, tc.first, tc.second, vr)
		}
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

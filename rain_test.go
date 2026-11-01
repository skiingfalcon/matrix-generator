package main

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

func newTestRain(width, height int, st style) *rain {
	return newRain(width, height, st, rand.New(rand.NewPCG(1, 2)))
}

var green = style{colour: 32}

// animate runs n frames, drawing each one on s.
func animate(r *rain, s *screen, n int) {
	s.t.Helper()
	for range n {
		r.tick()
		s.write(r.render(nil))
	}
}

// checkMatchesModel fails the test unless the screen shows exactly the
// visible part of every column: no leftover fragments, nothing missing.
func checkMatchesModel(t *testing.T, r *rain, s *screen) {
	t.Helper()
	want := make([][]rune, s.height)
	for y := range want {
		want[y] = make([]rune, s.width)
	}
	for x, c := range r.columns {
		if c.delay > 0 {
			continue
		}
		for i, ch := range c.chars {
			if y := c.head - i; y >= 0 && y < s.height {
				want[y][x] = ch
			}
		}
	}
	for y := range s.height {
		for x := range s.width {
			got := s.at(x, y)
			switch {
			case want[y][x] == 0 && !got.blank():
				t.Fatalf("leftover %q at column %d, row %d", got.ch, x, y)
			case want[y][x] != 0 && got.ch != want[y][x]:
				t.Fatalf("column %d, row %d shows %q, want %q", x, y, got.ch, want[y][x])
			}
		}
	}
}

// trails returns every trail on the screen as a list of cells from the
// bottom up. A trail is a run of drawn cells in one column.
func trails(s *screen) [][]cell {
	var out [][]cell
	for x := range s.width {
		var run []cell
		for y := s.height - 1; y >= -1; y-- {
			if y >= 0 && !s.at(x, y).blank() {
				run = append(run, s.at(x, y))
				continue
			}
			if len(run) > 0 {
				out = append(out, run)
				run = nil
			}
		}
	}
	return out
}

func TestRenderMatchesModel(t *testing.T) {
	for _, size := range [][2]int{{1, 1}, {2, 1}, {1, 2}, {3, 5}, {80, 24}, {200, 60}} {
		for _, kana := range []bool{false, true} {
			t.Run(fmt.Sprintf("%dx%d/katakana=%v", size[0], size[1], kana), func(t *testing.T) {
				r := newTestRain(size[0], size[1], style{colour: 32, katakana: kana})
				s := newScreen(t, size[0], size[1])
				s.write(r.render(nil))
				for range 20 {
					animate(r, s, 25)
					checkMatchesModel(t, r, s)
				}
			})
		}
	}
}

func TestResizeStaysOnScreen(t *testing.T) {
	r := newTestRain(80, 24, green)
	s := newScreen(t, 80, 24)
	animate(r, s, 200)
	for _, size := range [][2]int{{40, 10}, {1, 1}, {120, 40}, {120, 5}, {3, 60}, {80, 24}} {
		s.resize(size[0], size[1])
		r.resize(size[0], size[1])
		s.write(r.render(nil))
		checkMatchesModel(t, r, s)
		animate(r, s, 150)
		checkMatchesModel(t, r, s)
	}
}

func TestResizeCarriesOn(t *testing.T) {
	r := newTestRain(80, 24, green)
	s := newScreen(t, 80, 24)
	animate(r, s, 200)
	before := make([][]cell, 24)
	for y := range before {
		before[y] = slices.Clone(s.cells[y])
	}

	// Shrinking clears the screen and redraws the columns that are left
	// straight away, unchanged.
	s.resize(50, 24)
	r.resize(50, 24)
	s.write(r.render(nil))
	drawn := 0
	for y := range 24 {
		for x := range 50 {
			got, want := s.at(x, y), before[y][x]
			if got.blank() != want.blank() || (!got.blank() && got.ch != want.ch) {
				t.Fatalf("column %d, row %d changed from %q to %q on resize", x, y, want.ch, got.ch)
			}
			if !s.at(x, y).blank() {
				drawn++
			}
		}
	}
	if drawn == 0 {
		t.Fatal("screen is empty after resize; the animation should carry on")
	}
}

func TestHeadIsBrightWhite(t *testing.T) {
	for _, st := range []style{green, {colour: 31}, {colour: 37}, {rainbow: true}} {
		r := newTestRain(120, 40, st)
		s := newScreen(t, 120, 40)
		animate(r, s, 300)
		checked := 0
		for x, c := range r.columns {
			if c.delay > 0 || c.head < 0 || c.head >= s.height {
				continue
			}
			if got := s.at(x, c.head); got.fg != 37 || !got.bold || got.dim {
				t.Fatalf("%+v: head at column %d is %+v, want bold white", st, x, got)
			}
			checked++
		}
		if checked == 0 {
			t.Fatalf("%+v: no heads on screen to check", st)
		}
	}
}

func TestTrailColourAndFade(t *testing.T) {
	r := newTestRain(120, 40, style{colour: 31})
	s := newScreen(t, 120, 40)
	animate(r, s, 300)

	// Brightness of a trail cell: 2 bold, 1 normal, 0 dim.
	level := func(c cell) int {
		switch {
		case c.bold && c.dim:
			t.Fatalf("cell is both bold and dim: %+v", c)
		case c.bold:
			return 2
		case c.dim:
			return 0
		}
		return 1
	}
	checked := 0
	for x, c := range r.columns {
		// Check whole trails with the head on screen and long enough to
		// show all three steps.
		top := c.head - c.length + 1
		if c.delay > 0 || c.length < 6 || top < 0 || c.head >= s.height {
			continue
		}
		seen := map[int]bool{}
		prev := 2
		for y := c.head - 1; y >= top; y-- {
			cl := s.at(x, y)
			if cl.fg != 31 {
				t.Fatalf("trail cell at column %d, row %d has colour %d, want 31", x, y, cl.fg)
			}
			l := level(cl)
			if l > prev {
				t.Fatalf("column %d gets brighter further from the head", x)
			}
			prev = l
			seen[l] = true
		}
		if len(seen) != 3 {
			t.Fatalf("column %d (length %d) does not fade through bold, normal and dim: %v", x, c.length, seen)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no complete trails on screen to check")
	}
}

func TestRainbow(t *testing.T) {
	colourSet := func(st style) map[int]bool {
		r := newTestRain(120, 40, st)
		s := newScreen(t, 120, 40)
		animate(r, s, 300)
		set := map[int]bool{}
		for _, trail := range trails(s) {
			for _, c := range trail[1:] {
				set[c.fg] = true
			}
		}
		return set
	}
	if got := colourSet(green); len(got) != 1 || !got[32] {
		t.Errorf("green trails use colours %v, want only 32", got)
	}
	got := colourSet(style{rainbow: true})
	if len(got) < 5 {
		t.Errorf("rainbow trails use colours %v, want at least 5", got)
	}
	if got[30] {
		t.Error("rainbow trails should not use black")
	}
}

func TestCharacters(t *testing.T) {
	tests := []struct {
		name string
		st   style
		ok   func(rune) bool
	}{
		{"ascii", green, func(r rune) bool { return strings.ContainsRune(asciiChars, r) }},
		{"katakana", style{colour: 32, katakana: true}, func(r rune) bool { return r >= 0xFF66 && r <= 0xFF9D }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRain(80, 24, tt.st)
			s := newScreen(t, 80, 24)
			animate(r, s, 300)
			for _, trail := range trails(s) {
				for _, c := range trail {
					if !tt.ok(c.ch) {
						t.Fatalf("unexpected character %q (U+%04X)", c.ch, c.ch)
					}
				}
			}
		})
	}
}

func TestRenderSkipsColumnsThatDidNotMove(t *testing.T) {
	r := newTestRain(80, 24, green)
	s := newScreen(t, 80, 24)
	animate(r, s, 100)
	if out := r.render(nil); len(out) != 0 {
		t.Fatalf("second render with no tick in between drew %d bytes, want 0", len(out))
	}
}

func benchmarkRain(b *testing.B) *rain {
	r := newTestRain(200, 60, green)
	for range 500 {
		r.tick()
		r.render(nil)
	}
	return r
}

func BenchmarkTickRender(b *testing.B) {
	r := benchmarkRain(b)
	var frame []byte
	total := 0
	b.ReportAllocs()
	for b.Loop() {
		r.tick()
		frame = r.render(frame[:0])
		total += len(frame)
	}
	b.ReportMetric(float64(total)/float64(b.N), "bytes/frame")
}

func BenchmarkFullRedraw(b *testing.B) {
	r := benchmarkRain(b)
	var frame []byte
	b.ReportAllocs()
	for b.Loop() {
		r.clear = true
		frame = r.render(frame[:0])
	}
	b.ReportMetric(float64(len(frame)), "bytes/frame")
}

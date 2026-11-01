package main

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// screen is a tiny terminal emulator for tests. It understands exactly the
// escape sequences ccmatrix draws with and fails the test on anything else,
// including drawing outside the screen.
type screen struct {
	t             *testing.T
	width, height int
	cells         [][]cell
	x, y          int
	pen           cell // attributes for the next character
}

type cell struct {
	ch   rune // 0 for a cell nothing has been drawn in
	fg   int  // 0 for the default colour
	bold bool
	dim  bool
}

func (c cell) blank() bool { return c.ch == 0 || c.ch == ' ' }

func newScreen(t *testing.T, width, height int) *screen {
	s := &screen{t: t}
	s.resize(width, height)
	return s
}

// resize changes the size the way a terminal does: what is still in view is
// kept and new cells are empty.
func (s *screen) resize(width, height int) {
	cells := make([][]cell, height)
	for y := range cells {
		cells[y] = make([]cell, width)
		if y < len(s.cells) {
			copy(cells[y], s.cells[y])
		}
	}
	s.width, s.height, s.cells = width, height, cells
}

func (s *screen) write(b []byte) {
	s.t.Helper()
	for len(b) > 0 {
		if b[0] != keyEsc {
			r, n := utf8.DecodeRune(b)
			if r == utf8.RuneError {
				s.t.Fatalf("invalid UTF-8 in output: %q", b)
			}
			if s.x < 0 || s.x >= s.width || s.y < 0 || s.y >= s.height {
				s.t.Fatalf("drew %q at column %d, row %d, outside a %dx%d screen", r, s.x, s.y, s.width, s.height)
			}
			c := s.pen
			c.ch = r
			s.cells[s.y][s.x] = c
			s.x++
			b = b[n:]
			continue
		}
		if len(b) < 2 || b[1] != '[' {
			s.t.Fatalf("unexpected escape sequence in output: %q", b)
		}
		end := 2
		for end < len(b) && (b[end] < 0x40 || b[end] > 0x7E) {
			end++
		}
		if end == len(b) {
			s.t.Fatalf("unterminated escape sequence: %q", b)
		}
		s.csi(string(b[2:end]), b[end])
		b = b[end+1:]
	}
}

func (s *screen) csi(params string, final byte) {
	s.t.Helper()
	var nums []int
	if params != "" {
		for _, p := range strings.Split(params, ";") {
			n, err := strconv.Atoi(p)
			if err != nil {
				s.t.Fatalf("bad parameters %q for %q", params, final)
			}
			nums = append(nums, n)
		}
	}
	switch final {
	case 'H':
		if len(nums) != 2 {
			s.t.Fatalf("cursor move needs row and column, got %q", params)
		}
		s.y, s.x = nums[0]-1, nums[1]-1
		if s.x < 0 || s.x >= s.width || s.y < 0 || s.y >= s.height {
			s.t.Fatalf("cursor moved to column %d, row %d, outside a %dx%d screen", s.x, s.y, s.width, s.height)
		}
	case 'J':
		if params != "2" {
			s.t.Fatalf("unexpected erase %q", params)
		}
		for y := range s.cells {
			clear(s.cells[y])
		}
	case 'm':
		if len(nums) == 0 {
			nums = []int{0}
		}
		for _, n := range nums {
			switch {
			case n == 0:
				s.pen = cell{}
			case n == 1:
				s.pen.bold = true
			case n == 2:
				s.pen.dim = true
			case n >= 30 && n <= 37:
				s.pen.fg = n
			default:
				s.t.Fatalf("unexpected SGR parameter %d", n)
			}
		}
	default:
		s.t.Fatalf("unexpected escape sequence %q%c", params, final)
	}
}

// at returns the cell at column x, row y.
func (s *screen) at(x, y int) cell { return s.cells[y][x] }

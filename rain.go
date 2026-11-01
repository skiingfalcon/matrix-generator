package main

import (
	"math/rand/v2"
	"unicode/utf8"
)

const (
	asciiChars    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	katakanaFirst = 0xFF66 // half-width katakana WO
	katakanaLast  = 0xFF9D // half-width katakana N
)

// column is one vertical stream of falling characters. chars[0] is the head
// and chars[t] is drawn t rows above it, so the whole trail moves with the
// head.
type column struct {
	head      int // row of the head; negative while still above the screen
	length    int // trail length in rows, including the head
	speed     int // frames between advances
	countdown int // frames left until the next advance
	delay     int // frames to wait before starting
	chars     []rune
	dirty     bool // moved since it was last drawn
}

// rain is a screen full of columns. It is not safe for concurrent use.
type rain struct {
	width, height int
	columns       []column
	style         style
	rng           *rand.Rand
	sgr           sgrTable
	clear         bool // the next render must clear the screen and redraw everything
}

func newRain(width, height int, st style, rng *rand.Rand) *rain {
	r := &rain{style: st, rng: rng, sgr: newSGRTable()}
	r.resize(width, height)
	return r
}

// resize fits the rain to a new screen size. Columns that are still on screen
// keep falling; new columns start after a random delay. The next render
// clears the screen and redraws it in full.
func (r *rain) resize(width, height int) {
	r.width = max(width, 1)
	r.height = max(height, 1)
	if r.width < len(r.columns) {
		r.columns = r.columns[:r.width]
	}
	for len(r.columns) < r.width {
		r.columns = append(r.columns, r.newColumn(true))
	}
	r.clear = true
}

// newColumn returns a column that starts above the screen. If staggered is
// true it also waits a random number of frames, so that a fresh screen does
// not start with every column in step.
func (r *rain) newColumn(staggered bool) column {
	length := min(3+r.rng.IntN(max(r.height/2, 1)), r.height)
	c := column{
		head:   -1 - r.rng.IntN(r.height),
		length: length,
		speed:  1 + r.rng.IntN(4),
		chars:  make([]rune, length),
		dirty:  true,
	}
	c.countdown = c.speed
	if staggered {
		c.delay = r.rng.IntN(r.height * 2)
	}
	for i := range c.chars {
		c.chars[i] = r.randomChar()
	}
	return c
}

func (r *rain) randomChar() rune {
	if r.style.katakana {
		return rune(katakanaFirst + r.rng.IntN(katakanaLast-katakanaFirst+1))
	}
	return rune(asciiChars[r.rng.IntN(len(asciiChars))])
}

// tick advances the animation by one frame. Call render after every tick:
// render blanks only the one cell each column has just left.
func (r *rain) tick() {
	for i := range r.columns {
		c := &r.columns[i]
		if c.delay > 0 {
			c.delay--
			continue
		}
		c.countdown--
		if c.countdown > 0 {
			continue
		}
		c.countdown = c.speed
		c.head++
		c.dirty = true

		// Flicker: now and then change one character in the trail.
		if r.rng.IntN(3) == 0 {
			c.chars[r.rng.IntN(len(c.chars))] = r.randomChar()
		}
		// Shift the trail down one place and give the head a new character.
		copy(c.chars[1:], c.chars)
		c.chars[0] = r.randomChar()

		if c.head-c.length >= r.height {
			*c = r.newColumn(false)
		}
	}
}

// render appends the escape sequences that bring the screen up to date to dst
// and returns the result. Only columns that moved since the last render are
// drawn, unless the screen has to be cleared after a resize.
func (r *rain) render(dst []byte) []byte {
	full := r.clear
	if full {
		dst = append(dst, clearScreen...)
		r.clear = false
	}
	current := "" // the SGR sequence in effect, to avoid sending it twice
	for x := range r.columns {
		c := &r.columns[x]
		if c.delay > 0 || (!c.dirty && !full) {
			continue
		}
		c.dirty = false
		for t := range c.length {
			y := c.head - t
			if y < 0 || y >= r.height {
				continue
			}
			dst = moveTo(dst, x, y)
			if s := r.cellSGR(t, c.length); s != current {
				dst = append(dst, s...)
				current = s
			}
			dst = utf8.AppendRune(dst, c.chars[t])
		}
		// Blank the cell the end of the trail has just left. A space only
		// shows the background, which is never changed, so the SGR state
		// does not matter here.
		if y := c.head - c.length; y >= 0 && y < r.height {
			dst = moveTo(dst, x, y)
			dst = append(dst, ' ')
		}
	}
	if current != "" {
		dst = append(dst, sgrReset...)
	}
	return dst
}

// cellSGR returns the SGR sequence for the cell t rows above the head of a
// trail of the given length: a bright white head, then the trail colour
// fading from bold to normal to dim.
func (r *rain) cellSGR(t, length int) string {
	if t == 0 {
		return sgrHead
	}
	code := r.style.colour
	if r.style.rainbow {
		code = rainbowCodes[t%len(rainbowCodes)]
	}
	switch {
	case t < length/3:
		return r.sgr[code-30][fadeBright]
	case t < 2*length/3:
		return r.sgr[code-30][fadeNormal]
	default:
		return r.sgr[code-30][fadeDim]
	}
}

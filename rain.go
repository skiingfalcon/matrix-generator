package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
)

// Supported trail colours and their ANSI foreground codes.
var colourCodes = map[string]int{
	"black":   30,
	"red":     31,
	"green":   32,
	"yellow":  33,
	"blue":    34,
	"magenta": 35,
	"cyan":    36,
	"white":   37,
}

// SupportedColours returns the sorted list of colour names for error messages.
func SupportedColours() []string {
	return []string{"black", "blue", "cyan", "green", "magenta", "red", "white", "yellow"}
}

// ResolveColour looks up a colour name. Returns an error listing supported colours
// if the name is unknown.
func ResolveColour(name string) (int, error) {
	code, ok := colourCodes[strings.ToLower(name)]
	if !ok {
		return 0, fmt.Errorf("unknown colour %q; supported: %s",
			name, strings.Join(SupportedColours(), ", "))
	}
	return code, nil
}

// Rainbow palette used with -r (excludes black for visibility).
var rainbowCodes = []int{31, 33, 32, 36, 34, 35, 37}

const (
	asciiChars    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	katakanaStart = 0xFF66
	katakanaEnd   = 0xFF9D
)

// Config holds rendering options for the rain effect.
type Config struct {
	ColourCode int  // ANSI colour for trails
	Rainbow    bool // use rainbow colours instead of ColourCode
	Katakana   bool // use half-width katakana instead of ASCII
}

// Column is one vertical stream of falling characters.
type Column struct {
	head    int   // current head row (0-based); can be above or below the screen
	length  int   // trail length in rows
	speed   int   // frames between advances
	counter int   // countdown until next advance
	delay   int   // frames to wait before first appearance
	chars   []rune
}

// Rain owns every column and draws frames into a buffer.
type Rain struct {
	width, height int
	columns       []Column
	cfg           Config
	rng           *rand.Rand
}

// NewRain creates a rain field sized to the given terminal dimensions.
func NewRain(width, height int, cfg Config, rng *rand.Rand) *Rain {
	r := &Rain{cfg: cfg, rng: rng}
	r.Resize(width, height)
	return r
}

// Resize rebuilds columns to match a new terminal size.
func (r *Rain) Resize(width, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	r.width = width
	r.height = height
	r.columns = make([]Column, width)
	for i := range r.columns {
		r.resetColumn(i, true)
	}
}

// resetColumn re-initialises column i. If staggered is true the column waits a
// random delay before starting, which avoids a uniform first frame.
func (r *Rain) resetColumn(i int, staggered bool) {
	span := r.height / 2
	if span < 1 {
		span = 1
	}
	length := 3 + r.rng.Intn(span)
	if length > r.height {
		length = r.height
	}
	c := Column{
		head:   -1 - r.rng.Intn(r.height), // start above the screen
		length: length,
		speed:  1 + r.rng.Intn(4),
		chars:  make([]rune, length),
	}
	for j := range c.chars {
		c.chars[j] = r.randomChar()
	}
	if staggered {
		c.delay = r.rng.Intn(r.height * 2)
	}
	c.counter = c.speed
	r.columns[i] = c
}

func (r *Rain) randomChar() rune {
	if r.cfg.Katakana {
		return rune(katakanaStart + r.rng.Intn(katakanaEnd-katakanaStart+1))
	}
	return rune(asciiChars[r.rng.Intn(len(asciiChars))])
}

// Tick advances every column by one frame.
func (r *Rain) Tick() {
	for i := range r.columns {
		c := &r.columns[i]
		if c.delay > 0 {
			c.delay--
			continue
		}
		c.counter--
		if c.counter > 0 {
			continue
		}
		c.counter = c.speed
		c.head++

		// Flicker: occasionally change a random trail character.
		if len(c.chars) > 0 && r.rng.Intn(3) == 0 {
			idx := r.rng.Intn(len(c.chars))
			c.chars[idx] = r.randomChar()
		}

		// Shift trail: new head char at front, drop oldest.
		newChar := r.randomChar()
		copy(c.chars[1:], c.chars[:len(c.chars)-1])
		c.chars[0] = newChar

		// Fully off the bottom: recycle.
		if c.head-c.length >= r.height {
			r.resetColumn(i, false)
		}
	}
}

// Render draws the current frame into buf.
func (r *Rain) Render(buf *bytes.Buffer) {
	buf.Reset()
	for col := 0; col < r.width; col++ {
		c := &r.columns[col]
		if c.delay > 0 {
			continue
		}
		for t := 0; t < c.length; t++ {
			row := c.head - t
			if row < 0 || row >= r.height {
				continue
			}
			ch := c.chars[t]
			buf.WriteString(MoveTo(col, row))
			if t == 0 {
				// Bright white head.
				buf.WriteString("\x1b[1;37m")
			} else {
				code := r.cfg.ColourCode
				if r.cfg.Rainbow {
					code = rainbowCodes[t%len(rainbowCodes)]
				}
				// Fade: bright near the head, dim further back.
				if t < c.length/3 {
					buf.WriteString(fmt.Sprintf("\x1b[1;%dm", code))
				} else if t < 2*c.length/3 {
					buf.WriteString(fmt.Sprintf("\x1b[%dm", code))
				} else {
					buf.WriteString(fmt.Sprintf("\x1b[2;%dm", code))
				}
			}
			buf.WriteRune(ch)
		}
		// Erase the cell just behind the trail.
		eraseRow := c.head - c.length
		if eraseRow >= 0 && eraseRow < r.height {
			buf.WriteString(MoveTo(col, eraseRow))
			buf.WriteString("\x1b[0m ")
		}
	}
	buf.WriteString("\x1b[0m")
}

// FrameDelayMs returns the duration between frames for a given -u value (0–10).
func FrameDelayMs(u int) int {
	return 10 + u*10
}

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"golang.org/x/term"
)

// Escape sequences. See the xterm control sequence documentation.
const (
	altScreenOn  = "\x1b[?1049h"
	altScreenOff = "\x1b[?1049l"
	cursorHide   = "\x1b[?25l"
	cursorShow   = "\x1b[?25h"
	clearScreen  = "\x1b[2J"
	mouseOn      = "\x1b[?1000h\x1b[?1006h" // report clicks, in SGR format
	mouseOff     = "\x1b[?1006l\x1b[?1000l"
	sgrReset     = "\x1b[0m"
	sgrHead      = "\x1b[0;1;37m" // bold white
)

// moveTo appends the sequence that moves the cursor to column x, row y
// (both 0-based) to dst.
func moveTo(dst []byte, x, y int) []byte {
	dst = append(dst, "\x1b["...)
	dst = strconv.AppendInt(dst, int64(y+1), 10)
	dst = append(dst, ';')
	dst = strconv.AppendInt(dst, int64(x+1), 10)
	return append(dst, 'H')
}

// Brightness steps along a trail.
const (
	fadeBright = iota
	fadeNormal
	fadeDim
)

// sgrTable holds the SGR sequence for each foreground colour (30-37, indexed
// from 0) at each brightness step. Each one starts with a reset, so no bold or
// dim attribute carries over from the cell drawn before it.
type sgrTable [8][3]string

func newSGRTable() sgrTable {
	var t sgrTable
	for i := range t {
		code := strconv.Itoa(30 + i)
		t[i][fadeBright] = "\x1b[0;1;" + code + "m"
		t[i][fadeNormal] = "\x1b[0;" + code + "m"
		t[i][fadeDim] = "\x1b[0;2;" + code + "m"
	}
	return t
}

// terminal is a terminal in raw mode showing the alternate screen.
type terminal struct {
	in    *os.File
	out   *os.File
	state *term.State
	mouse bool
}

// openTerminal puts in into raw mode and switches out to the alternate screen
// with the cursor hidden. If mouse is true, mouse clicks are reported as
// input. Call restore to undo all of this.
func openTerminal(in, out *os.File, mouse bool) (*terminal, error) {
	if !term.IsTerminal(int(in.Fd())) {
		return nil, errors.New("standard input is not a terminal")
	}
	if !term.IsTerminal(int(out.Fd())) {
		return nil, errors.New("standard output is not a terminal")
	}
	state, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		return nil, fmt.Errorf("enter raw mode: %w", err)
	}
	t := &terminal{in: in, out: out, state: state, mouse: mouse}

	setup := altScreenOn + cursorHide + clearScreen
	if mouse {
		setup += mouseOn
	}
	if _, err := io.WriteString(out, setup); err != nil {
		return nil, errors.Join(fmt.Errorf("set up screen: %w", err), t.restore())
	}
	return t, nil
}

// size returns the screen size in cells.
func (t *terminal) size() (width, height int, err error) {
	return term.GetSize(int(t.out.Fd()))
}

// restore puts the terminal back the way openTerminal found it: mouse
// reporting off, attributes reset, cursor shown, the original screen back and
// cooked mode on.
func (t *terminal) restore() error {
	teardown := sgrReset + cursorShow + altScreenOff
	if t.mouse {
		teardown = mouseOff + teardown
	}
	_, werr := io.WriteString(t.out, teardown)
	rerr := term.Restore(int(t.in.Fd()), t.state)
	if werr != nil {
		werr = fmt.Errorf("restore screen: %w", werr)
	}
	if rerr != nil {
		rerr = fmt.Errorf("leave raw mode: %w", rerr)
	}
	return errors.Join(werr, rerr)
}

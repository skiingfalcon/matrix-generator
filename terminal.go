package main

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/term"
)

// Terminal manages raw mode, alternate screen, cursor, and optional mouse reporting.
type Terminal struct {
	fd        int
	oldState  *term.State
	mouse     bool
	restored  bool
	mu        sync.Mutex
}

// NewTerminal puts the terminal into raw mode, switches to the alternate screen,
// and hides the cursor. If mouse is true, mouse reporting is enabled.
func NewTerminal(mouse bool) (*Terminal, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, fmt.Errorf("stdin is not a terminal")
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("enter raw mode: %w", err)
	}

	t := &Terminal{fd: fd, oldState: oldState, mouse: mouse}

	// Enter alternate screen, hide cursor, clear screen, move home.
	fmt.Fprint(os.Stdout, "\x1b[?1049h\x1b[?25l\x1b[2J\x1b[H")
	if mouse {
		// X10 mouse reporting + SGR extended mouse reporting.
		fmt.Fprint(os.Stdout, "\x1b[?1000h\x1b[?1006h")
	}

	return t, nil
}

// Size returns the current terminal width and height in cells.
func (t *Terminal) Size() (width, height int, err error) {
	return term.GetSize(t.fd)
}

// Restore returns the terminal to its previous state. Safe to call more than once.
func (t *Terminal) Restore() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.restored {
		return
	}
	t.restored = true

	if t.mouse {
		fmt.Fprint(os.Stdout, "\x1b[?1006l\x1b[?1000l")
	}
	// Reset attributes, show cursor, leave alternate screen.
	fmt.Fprint(os.Stdout, "\x1b[0m\x1b[?25h\x1b[?1049l")
	if t.oldState != nil {
		_ = term.Restore(t.fd, t.oldState)
	}
}

// Clear clears the screen and moves the cursor home.
func Clear() {
	fmt.Fprint(os.Stdout, "\x1b[2J\x1b[H")
}

// MoveTo returns an ANSI escape sequence that moves the cursor to (col, row).
// Columns and rows are 0-based; the escape sequence is 1-based.
func MoveTo(col, row int) string {
	return fmt.Sprintf("\x1b[%d;%dH", row+1, col+1)
}

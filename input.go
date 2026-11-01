package main

import (
	"bytes"
	"io"
)

const (
	keyEsc   = 0x1b
	keyCtrlC = 0x03
)

// readInput sends each chunk read from r on out, and closes out when it stops.
//
// It stops when a read fails (including io.EOF), or when done is closed and
// its current read returns. A blocked read on a terminal cannot be
// interrupted, so when the program quits this goroutine is usually still
// waiting for a key and simply ends with the process.
func readInput(r io.Reader, out chan<- []byte, done <-chan struct{}) {
	defer close(out)
	buf := make([]byte, 256)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			select {
			case out <- bytes.Clone(buf[:n]):
			case <-done:
				return
			}
		}
		if err != nil {
			return
		}
	}
}

// plainKeys returns the ordinary key presses in b, leaving out escape
// sequences such as arrow keys, function keys, Alt+key and mouse reports, so
// that the digits inside them are not taken as speed changes. A sequence split
// across two reads is not recognised; terminals send each one in one write.
func plainKeys(b []byte) []byte {
	var keys []byte
	for i := 0; i < len(b); i++ {
		if b[i] != keyEsc {
			keys = append(keys, b[i])
			continue
		}
		i = skipEscape(b, i)
	}
	return keys
}

// skipEscape returns the index of the last byte of the escape sequence that
// starts at b[i].
func skipEscape(b []byte, i int) int {
	if i+1 >= len(b) {
		return i // Esc on its own
	}
	switch b[i+1] {
	case '[': // CSI: parameters and intermediates, then a final byte 0x40-0x7E
		j := i + 2
		for j < len(b) && (b[j] < 0x40 || b[j] > 0x7E) {
			j++
		}
		if j == i+2 && j < len(b) && b[j] == 'M' {
			// Old-style mouse report: Esc [ M followed by three raw bytes.
			return min(j+3, len(b)-1)
		}
		return min(j, len(b)-1)
	case 'O': // SS3, e.g. F1-F4 and arrows in application mode: Esc O x
		return min(i+2, len(b)-1)
	default: // Alt+key: Esc x
		return i + 1
	}
}

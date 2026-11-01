package main

import (
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func TestPlainKeys(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"letter", "q", "q"},
		{"several keys in one read", "9q", "9q"},
		{"ctrl-c", "\x03", "\x03"},
		{"enter", "\r", "\r"},
		{"esc on its own", "\x1b", ""},
		{"up arrow", "\x1b[A", ""},
		{"delete", "\x1b[3~", ""},
		{"page up", "\x1b[5~", ""},
		{"F5", "\x1b[15~", ""},
		{"F1 in SS3 form", "\x1bOP", ""},
		{"shift+up", "\x1b[1;2A", ""},
		{"alt+5", "\x1b5", ""},
		{"SGR mouse click", "\x1b[<0;12;5M", ""},
		{"old-style mouse click", "\x1b[M 3%", ""},
		{"keys around a sequence", "a\x1b[Bb", "ab"},
		{"truncated sequence", "\x1b[1", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(plainKeys([]byte(tt.in))); got != tt.want {
				t.Errorf("plainKeys(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestReadInputForwardsEachRead(t *testing.T) {
	// OneByteReader returns one byte per read, so each byte is its own chunk.
	out := make(chan []byte)
	go readInput(iotest.OneByteReader(strings.NewReader("ab")), out, make(chan struct{}))
	var got []string
	for chunk := range out {
		got = append(got, string(chunk))
	}
	if strings.Join(got, ",") != "a,b" {
		t.Errorf("got chunks %q, want [a b]", got)
	}
}

func TestReadInputStopsWhenDone(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()
	out := make(chan []byte)
	done := make(chan struct{})
	go readInput(pr, out, done)

	// done is closed, so once this read returns readInput must stop
	// rather than wait for someone to take the chunk.
	close(done)
	pw.Write([]byte("x"))

	select {
	case _, ok := <-out:
		if !ok {
			return // stopped without sending
		}
		// Our receive was ready at the same moment, so the send could win
		// the select. Then readInput is back in a read, which only the
		// pipe closing can end.
		pw.Close()
		if _, ok := <-out; ok {
			t.Fatal("readInput sent a second chunk")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("readInput did not stop after done was closed")
	}
}

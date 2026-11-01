package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	screensaver := flag.Bool("s", false, "screensaver mode: exit on any key or mouse")
	updateStr := flag.String("u", "4", "update delay 0-10 (0=fastest, default 4)")
	colourName := flag.String("C", "green", "trail colour")
	rainbow := flag.Bool("r", false, "rainbow trail colours")
	katakana := flag.Bool("c", false, "use half-width katakana characters")

	flag.Parse()

	updateDelay, err := strconv.Atoi(*updateStr)
	if err != nil || updateDelay < 0 || updateDelay > 10 {
		fmt.Fprintf(os.Stderr, "ccmatrix: -u must be an integer from 0 to 10, got %q\n", *updateStr)
		os.Exit(1)
	}

	colourCode, err := ResolveColour(*colourName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccmatrix: %v\n", err)
		os.Exit(1)
	}

	cfg := Config{
		ColourCode: colourCode,
		Rainbow:    *rainbow,
		Katakana:   *katakana,
	}

	term, err := NewTerminal(*screensaver)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccmatrix: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore()

	// Restore on SIGINT / SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	winchCh := make(chan os.Signal, 1)
	signal.Notify(winchCh, syscall.SIGWINCH)

	width, height, err := term.Size()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccmatrix: get size: %v\n", err)
		os.Exit(1)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rain := NewRain(width, height, cfg, rng)

	keys := make(chan byte, 64)
	go readKeys(keys)

	speed := updateDelay
	ticker := time.NewTicker(time.Duration(FrameDelayMs(speed)) * time.Millisecond)
	defer ticker.Stop()

	var buf bytes.Buffer

	for {
		select {
		case <-sigCh:
			return

		case <-winchCh:
			w, h, err := term.Size()
			if err != nil {
				continue
			}
			Clear()
			rain.Resize(w, h)

		case b := <-keys:
			if *screensaver {
				return
			}
			switch {
			case b == 'q' || b == 'Q':
				return
			case b >= '0' && b <= '9':
				speed = int(b - '0')
				ticker.Reset(time.Duration(FrameDelayMs(speed)) * time.Millisecond)
			}
			// Drain any remaining buffered escape / mouse bytes in screensaver
			// is already handled above; for normal mode ignore extras.
			drainKeys(keys)

		case <-ticker.C:
			rain.Tick()
			rain.Render(&buf)
			os.Stdout.Write(buf.Bytes())
		}
	}
}

// readKeys reads raw bytes from stdin and forwards them on ch.
func readKeys(ch chan<- byte) {
	buf := make([]byte, 64)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}
		for i := 0; i < n; i++ {
			select {
			case ch <- buf[i]:
			default:
			}
		}
	}
}

// drainKeys empties the key channel without blocking.
func drainKeys(ch <-chan byte) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

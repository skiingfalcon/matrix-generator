// Command ccmatrix fills the terminal with falling text, like the film The
// Matrix. See README.md for the options.
package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := parseConfig(os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		os.Exit(2) // parseConfig has already reported the problem
	}
	if err := run(cfg, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "ccmatrix: %v\n", err)
		os.Exit(1)
	}
}

// run animates the rain on the terminal until the user quits or a signal
// arrives. The terminal is restored before run returns, whatever happens.
func run(cfg config, stdin, stdout *os.File) (err error) {
	t, err := openTerminal(stdin, stdout, cfg.screensaver)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, t.restore())
	}()

	width, height, err := t.size()
	if err != nil {
		return fmt.Errorf("get terminal size: %w", err)
	}
	r := newRain(width, height, cfg.style, rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)))

	// In raw mode Ctrl-C is an ordinary key, so these come from elsewhere,
	// such as kill or the terminal closing.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(quit)

	resized := make(chan os.Signal, 1)
	signal.Notify(resized, syscall.SIGWINCH)
	defer signal.Stop(resized)

	done := make(chan struct{})
	defer close(done)
	input := make(chan []byte)
	go readInput(stdin, input, done)

	ticker := time.NewTicker(frameDelay(cfg.delay))
	defer ticker.Stop()

	var frame []byte
	draw := func() error {
		frame = r.render(frame[:0])
		if _, err := stdout.Write(frame); err != nil {
			return fmt.Errorf("draw: %w", err)
		}
		return nil
	}

	for {
		select {
		case <-quit:
			return nil

		case <-resized:
			width, height, err := t.size()
			if err != nil {
				return fmt.Errorf("get terminal size: %w", err)
			}
			r.resize(width, height)
			if err := draw(); err != nil {
				return err
			}

		case b, ok := <-input:
			if !ok {
				return errors.New("standard input closed")
			}
			if cfg.screensaver {
				return nil
			}
			for _, k := range plainKeys(b) {
				switch {
				case k == 'q' || k == 'Q' || k == keyCtrlC:
					return nil
				case k >= '0' && k <= '9':
					ticker.Reset(frameDelay(int(k - '0')))
				}
			}

		case <-ticker.C:
			r.tick()
			if err := draw(); err != nil {
				return err
			}
		}
	}
}

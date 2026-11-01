package main

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// config is everything the command line controls.
type config struct {
	screensaver bool // exit on any key or mouse click
	delay       int  // -u value, 0 (fastest) to maxDelay
	style       style
}

// style is the part of config that changes how the rain looks.
type style struct {
	colour   int  // ANSI foreground code for trails (30-37)
	rainbow  bool // colour trails from rainbowCodes instead of colour
	katakana bool // draw half-width katakana instead of letters and digits
}

const (
	defaultDelay = 4
	maxDelay     = 10
)

// colours lists the supported -C values in the order they are reported.
var colours = []struct {
	name string
	code int
}{
	{"green", 32},
	{"red", 31},
	{"blue", 34},
	{"white", 37},
	{"yellow", 33},
	{"cyan", 36},
	{"magenta", 35},
	{"black", 30},
}

// rainbowCodes are the trail colours used by -r. Black is left out because it
// is invisible on most backgrounds.
var rainbowCodes = []int{31, 33, 32, 36, 34, 35, 37}

// frameDelay is the time between frames for a -u value.
func frameDelay(u int) time.Duration {
	return time.Duration(10+u*10) * time.Millisecond
}

// delayFlag is the -u flag. Set rejects anything that is not an integer from
// 0 to maxDelay, so the flag package reports the problem and the usage.
type delayFlag int

func (d *delayFlag) String() string { return strconv.Itoa(int(*d)) }

func (d *delayFlag) Set(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n > maxDelay {
		return fmt.Errorf("must be an integer from 0 to %d", maxDelay)
	}
	*d = delayFlag(n)
	return nil
}

// colourFlag is the -C flag. It holds the ANSI code of the chosen colour.
type colourFlag int

func (c *colourFlag) String() string {
	for _, col := range colours {
		if col.code == int(*c) {
			return col.name
		}
	}
	return ""
}

func (c *colourFlag) Set(s string) error {
	for _, col := range colours {
		if strings.EqualFold(s, col.name) {
			*c = colourFlag(col.code)
			return nil
		}
	}
	names := make([]string, len(colours))
	for i, col := range colours {
		names[i] = col.name
	}
	return fmt.Errorf("unknown colour; supported colours are %s", strings.Join(names, ", "))
}

// parseConfig parses the command line arguments (without the program name).
// Problems are reported to stderr, along with the usage, by parseConfig
// itself; the returned error only tells the caller to stop. -h returns
// flag.ErrHelp.
func parseConfig(args []string, stderr io.Writer) (config, error) {
	fs := flag.NewFlagSet("ccmatrix", flag.ContinueOnError)
	fs.SetOutput(stderr)

	delay := delayFlag(defaultDelay)
	colour := colourFlag(32)
	var cfg config
	fs.BoolVar(&cfg.screensaver, "s", false, "screensaver mode: exit on any key or mouse click")
	fs.Var(&delay, "u", fmt.Sprintf("update `delay` from 0 (fastest) to %d", maxDelay))
	fs.Var(&colour, "C", "trail `colour`: green, red, blue, white, yellow, cyan, magenta or black")
	fs.BoolVar(&cfg.style.rainbow, "r", false, "rainbow trail colours")
	fs.BoolVar(&cfg.style.katakana, "c", false, "use half-width katakana characters")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	if fs.NArg() > 0 {
		err := fmt.Errorf("unexpected argument %q", fs.Arg(0))
		fmt.Fprintln(stderr, err)
		fs.Usage()
		return config{}, err
	}
	cfg.delay = int(delay)
	cfg.style.colour = int(colour)
	return cfg, nil
}

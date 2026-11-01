package main

import (
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		args []string
		want config
	}{
		{nil, config{delay: 4, style: style{colour: 32}}},
		{[]string{"-s"}, config{screensaver: true, delay: 4, style: style{colour: 32}}},
		{[]string{"-u", "0"}, config{delay: 0, style: style{colour: 32}}},
		{[]string{"-u", "10"}, config{delay: 10, style: style{colour: 32}}},
		{[]string{"-C", "red"}, config{delay: 4, style: style{colour: 31}}},
		{[]string{"-C", "CYAN"}, config{delay: 4, style: style{colour: 36}}},
		{[]string{"-C", "black"}, config{delay: 4, style: style{colour: 30}}},
		{[]string{"-r"}, config{delay: 4, style: style{colour: 32, rainbow: true}}},
		{[]string{"-c", "-C", "cyan", "-u", "2"}, config{delay: 2, style: style{colour: 36, katakana: true}}},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			var stderr strings.Builder
			got, err := parseConfig(tt.args, &stderr)
			if err != nil {
				t.Fatalf("unexpected error %v; stderr:\n%s", err, stderr.String())
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
			if stderr.Len() > 0 {
				t.Errorf("unexpected output on stderr:\n%s", stderr.String())
			}
		})
	}
}

func TestParseConfigErrors(t *testing.T) {
	tests := []struct {
		args     []string
		wantText []string // must all appear on stderr
	}{
		{[]string{"-u", "11"}, []string{"-u", "0 to 10"}},
		{[]string{"-u", "-1"}, []string{"-u", "0 to 10"}},
		{[]string{"-u", "fast"}, []string{"-u", "0 to 10"}},
		{[]string{"-u", "4.5"}, []string{"-u", "0 to 10"}},
		{[]string{"-C", "purple"}, []string{"purple", "green", "red", "blue", "white", "yellow", "cyan", "magenta", "black"}},
		{[]string{"-x"}, []string{"-x"}},
		{[]string{"extra"}, []string{"extra"}},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			var stderr strings.Builder
			_, err := parseConfig(tt.args, &stderr)
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, s := range tt.wantText {
				if !strings.Contains(stderr.String(), s) {
					t.Errorf("stderr does not mention %q:\n%s", s, stderr.String())
				}
			}
			if !strings.Contains(stderr.String(), "Usage") {
				t.Errorf("stderr does not show the usage:\n%s", stderr.String())
			}
		})
	}
}

func TestParseConfigHelp(t *testing.T) {
	var stderr strings.Builder
	_, err := parseConfig([]string{"-h"}, &stderr)
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("got %v, want flag.ErrHelp", err)
	}
	for _, s := range []string{"-u", "(default 4)", "-C", "(default green)"} {
		if !strings.Contains(stderr.String(), s) {
			t.Errorf("usage does not mention %q:\n%s", s, stderr.String())
		}
	}
}

func TestFrameDelay(t *testing.T) {
	for u := 1; u <= maxDelay; u++ {
		if frameDelay(u) <= frameDelay(u-1) {
			t.Errorf("-u %d (%v) is not slower than -u %d (%v)", u, frameDelay(u), u-1, frameDelay(u-1))
		}
	}
	if frameDelay(0) <= 0 {
		t.Errorf("-u 0 gives %v; a zero ticker period panics", frameDelay(0))
	}
}

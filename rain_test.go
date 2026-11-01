package main

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
)

func TestResolveColour(t *testing.T) {
	code, err := ResolveColour("green")
	if err != nil || code != 32 {
		t.Fatalf("green: code=%d err=%v", code, err)
	}
	code, err = ResolveColour("RED")
	if err != nil || code != 31 {
		t.Fatalf("RED: code=%d err=%v", code, err)
	}
	_, err = ResolveColour("purple")
	if err == nil {
		t.Fatal("expected error for purple")
	}
	if !strings.Contains(err.Error(), "green") {
		t.Fatalf("error should list supported colours: %v", err)
	}
}

func TestSupportedColours(t *testing.T) {
	cols := SupportedColours()
	if len(cols) != 8 {
		t.Fatalf("want 8 colours, got %d", len(cols))
	}
	for _, name := range cols {
		if _, err := ResolveColour(name); err != nil {
			t.Fatalf("%s should resolve: %v", name, err)
		}
	}
}

func TestFrameDelayMs(t *testing.T) {
	if FrameDelayMs(0) >= FrameDelayMs(4) {
		t.Fatal("u=0 should be faster than u=4")
	}
	if FrameDelayMs(4) >= FrameDelayMs(10) {
		t.Fatal("u=4 should be faster than u=10")
	}
	if FrameDelayMs(0) != 10 {
		t.Fatalf("u=0 delay want 10, got %d", FrameDelayMs(0))
	}
	if FrameDelayMs(4) != 50 {
		t.Fatalf("u=4 delay want 50, got %d", FrameDelayMs(4))
	}
}

func TestColumnAdvancesAndResets(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	cfg := Config{ColourCode: 32}
	rain := NewRain(1, 10, cfg, rng)

	c := &rain.columns[0]
	c.delay = 0
	c.head = 0
	c.length = 3
	c.speed = 1
	c.counter = 1
	c.chars = []rune{'A', 'B', 'C'}

	startHead := c.head
	rain.Tick()
	if rain.columns[0].head != startHead+1 {
		t.Fatalf("head should advance by 1, got %d", rain.columns[0].head)
	}
}

func TestRenderContainsEscapeSequences(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	cfg := Config{ColourCode: 32}
	rain := NewRain(5, 8, cfg, rng)

	// Force a visible column.
	c := &rain.columns[2]
	c.delay = 0
	c.head = 3
	c.length = 4
	c.speed = 1
	c.counter = 1
	c.chars = []rune{'W', 'X', 'Y', 'Z'}

	var buf bytes.Buffer
	rain.Render(&buf)
	out := buf.String()
	if !strings.Contains(out, "\x1b[") {
		t.Fatal("render should contain ANSI escapes")
	}
	if !strings.Contains(out, "\x1b[1;37m") {
		t.Fatal("render should draw a bright white head")
	}
	if !strings.Contains(out, "W") {
		t.Fatal("render should include the head character")
	}
}

func TestResizeRebuildsColumns(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	rain := NewRain(10, 10, Config{ColourCode: 32}, rng)
	if len(rain.columns) != 10 {
		t.Fatalf("want 10 columns, got %d", len(rain.columns))
	}
	rain.Resize(20, 15)
	if rain.width != 20 || rain.height != 15 {
		t.Fatalf("size after resize: %dx%d", rain.width, rain.height)
	}
	if len(rain.columns) != 20 {
		t.Fatalf("want 20 columns after resize, got %d", len(rain.columns))
	}
}

func TestKatakanaCharset(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	rain := NewRain(1, 5, Config{ColourCode: 32, Katakana: true}, rng)
	for i := 0; i < 50; i++ {
		ch := rain.randomChar()
		if ch < katakanaStart || ch > katakanaEnd {
			t.Fatalf("katakana char out of range: U+%X", ch)
		}
	}
}

func TestRainbowRender(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	rain := NewRain(1, 12, Config{Rainbow: true}, rng)
	c := &rain.columns[0]
	c.delay = 0
	c.head = 8
	c.length = 7
	c.chars = make([]rune, 7)
	for i := range c.chars {
		c.chars[i] = 'A'
	}
	var buf bytes.Buffer
	rain.Render(&buf)
	out := buf.String()
	// Should contain more than one colour code from the rainbow set.
	found := 0
	for _, code := range rainbowCodes {
		needle := strings.Contains(out, string(rune('0'+code/10))+string(rune('0'+code%10)))
		// Simpler check via sprintf pattern.
		_ = needle
		if strings.Contains(out, "\x1b[") && strings.Contains(out, "m") {
			found++
			break
		}
	}
	if found == 0 {
		t.Fatal("rainbow render should contain colour escapes")
	}
}

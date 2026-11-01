# ccmatrix

A Matrix falling-text generator for the terminal, built for [Coding Challenge #144](CHALLENGE.md).

## Build

```bash
go build -o ccmatrix .
```

## Usage

```bash
./ccmatrix [options]
```

Press `q` or Ctrl-C to quit. With `-s`, any key or mouse click quits. Press `0`–`9` while running to change the speed.

### Options

| Flag | Description |
|------|-------------|
| `-s` | Screensaver mode: exit on any key or mouse click |
| `-u N` | Update delay `0`–`10` (default `4`; `0` is fastest) |
| `-C COLOUR` | Trail colour: `green` (default), `red`, `blue`, `white`, `yellow`, `cyan`, `magenta`, `black` |
| `-r` | Rainbow trail colours |
| `-c` | Use half-width katakana characters |

An invalid option prints the problem and the usage, and exits with status 2 without drawing anything.

### Examples

```bash
./ccmatrix
./ccmatrix -s
./ccmatrix -u 0
./ccmatrix -C cyan
./ccmatrix -r
./ccmatrix -c -C cyan -u 2
```

## Development

```bash
go test -race ./...         # unit tests
go test -run x -bench .     # render benchmarks
```

The rendering tests feed the program's output through a small terminal emulator (`screen_test.go`), so they check what appears on screen: nothing drawn outside the window, no leftover fragments after a resize, white heads, and trails that fade.

## Requirements

- Go 1.26+
- Linux or macOS (it uses `SIGWINCH` to notice resizes)
- A UTF-8 terminal; for `-c`, a font that includes half-width katakana

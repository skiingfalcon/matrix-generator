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

Press `q` to quit (or any key / mouse click with `-s`). Press `0`–`9` while running to change speed.

### Options

| Flag | Description |
|------|-------------|
| `-s` | Screensaver mode: exit on any key or mouse input |
| `-u N` | Update delay `0`–`10` (default `4`; `0` is fastest) |
| `-C COLOUR` | Trail colour: `green` (default), `red`, `blue`, `white`, `yellow`, `cyan`, `magenta`, `black` |
| `-r` | Rainbow trail colours |
| `-c` | Use half-width katakana characters |

### Examples

```bash
./ccmatrix
./ccmatrix -s
./ccmatrix -u 0
./ccmatrix -C cyan
./ccmatrix -r
./ccmatrix -c -C cyan -u 2
```

## Requirements

- Go 1.21+
- A UTF-8 terminal (for `-c` katakana, a font that includes half-width katakana)

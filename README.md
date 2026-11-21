# Ranga-Tools

A command-line toolkit for managing, testing and analyzing UCI chess engines.

Register your engines once, then refer to them by name to run tournaments, analyze games, benchmark builds and prepare training data.

## Features

- **Engine management**: register, list, run, benchmark and remove UCI engines. Each engine is verified with the UCI handshake when you register it.
- **Tournaments**: run SPRT tests, head-to-head matches and self-play through [fastchess](https://github.com/Disservin/fastchess), driven by a JSON config file.
- **Game analysis**: analyze a player's games from a PGN with an engine. 
- **Viriformat tools**: preprocess `.vf` training data.

## Prerequisites

- [Go](https://go.dev/dl/) 1.27.4 or above
- [fastchess](https://github.com/Disservin/fastchess), required for tournaments only
- `make`

## Installation

```bash
git clone https://github.com/nchandur/ranga-tools.git
cd ranga-tools
make build
```

## Quick start

```bash
# 1. Register an engine
ranga-tools engine register --path ./builds/ranga-1.23

# 2. Tell the tool where fastchess lives (skip if it's on your PATH)
ranga-tools config set fastchess-path /usr/local/bin/fastchess

# 3. Run a tournament from a config file
ranga-tools tournament sprt --config sprt.json

# 4. Analyze a player's games
ranga-tools analyze --engine ranga-1.23 --pgn games.pgn --player HangMateIn
```

## Commands

| Command | What it does | Guide |
|---|---|---|
| `engine` | Register, list, deregister, run and benchmark engines | [docs/engine.md](docs/engine.md) |
| `tournament` | Run `sprt`, `match` and `selftest` tournaments via fastchess | [docs/tournament.md](docs/tournament.md) |
| `analyze` | Analyze a player's games from a PGN | [docs/analysis.md](docs/analysis.md) |
| `config` | Read and write tool settings such as the fastchess path | [docs/config.md](docs/config.md) |
| `viriformat` | Preprocess`.vf` training data | [docs/viriformat.md](docs/viriformat.md) |

Run `ranga-tools <command> --help` for the flags of any command.

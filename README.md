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
ranga-tools analyze --engine ranga-1.23 --pgn games.pgn --player niemann
```

## Commands

- `engine`: [Link](docs/engine.md)
- `tournament`: [Link](docs/tournament.md)
- `analyze`: [Link](docs/analysis.md)
- `config`: [Link](docs/config.md)
- `viriformat`: [Link](docs/viriformat.md)

Run `ranga-tools <command> --help` for the flags of any command.

# Ranga-Tools

Ranga-Tools is a command-line interface (CLI) designed for managing, configuring, and testing UCI (Universal Chess Interface) chess engines

## Features

- **Engine Management**: Register, list, and deregister UCI chess engines
- **Tournament Execution**: Run Sequential Probability Ratio Tests (SPRT) between engines to measure Elo improvements
- **Analysis**: Parses PGN game and provides an analysis report on performance and strength

## Installation and Prerequisites

- Go: `1.27.3` or above
- [Fastchess](https://github.com/Disservin/fastchess)
- `make`

To build binary: 
```bash
make build
```

Before running tournaments, configure the path to your Fastchess executable:

```bash
ranga-tools config set fastchess-path <path-to-fastchess>
```
> Note: Tournaments will not run if this path is not set

## Usage Guide

1. Managing Engines (`engine`): Refer [here](docs/engine.md)
2. Running Tournaments (`tournament`): Refer [here](docs/tournament.md)
3. Running Analysis (`analyze`): Refer [here](docs/analysis.md) 
4. Configuration Management (`config`): Refer [here](docs/config.md)

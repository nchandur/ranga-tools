# Usage Guide for the `tournament` Command

The `tournament` command runs engine-versus-engine tournaments through [fastchess](https://github.com/Disservin/fastchess). The tool builds the fastchess command line from a JSON config file and the engines you have registered, runs it, and streams fastchess's output to your terminal. Results come from fastchess itself.

```bash
ranga-tools tournament <match|sprt|selftest> --config <path-to-config> [--fastchess <path>]
```

## Prerequisites

- **fastchess installed.** See the [config guide](config.md) for how the binary is located.
- **Engines registered.** Every engine named in a config must be registered first (see the [engine guide](engine.md)).

## Common flags

- `--config`: Path to the config file. Default is `game-config.json` in the current directory.
- `--fastchess`: Path to the fastchess binary for this run. Overrides the stored setting.

Before fastchess starts, the full command line is printed to stderr as `running: ...`, so you can see exactly what was run.

## Match

A plain match between two different engines.

```bash
ranga-tools tournament match --config match.json
```

```json
{
  "engine1": { "name": "ranga-1.21", "options": { "Hash": "128", "Threads": "1" } },
  "engine2": { "name": "ranga-1.22", "options": { "Hash": "128", "Threads": "1" } },

  "nodes": 5000,
  "tc": "4+0.4",
  "rounds": 10,
  "games": 1,
  "concurrency": 7,

  "openings": { "file": "books/UHO_Lichess_4852_v1.epd" }
}
```

The two engines must be different. To play an engine against itself, use `selftest`. Here `tc` is set, so `nodes` is ignored (see the time control rules below).

## SPRT

A Sequential Probability Ratio Test between a release candidate (`rc`) and a baseline (`base`). The test stops early once the result is statistically decided.

```bash
ranga-tools tournament sprt --config sprt.json
```

```json
{
  "rc":   { "name": "ranga-1.21", "options": { "Hash": "128", "Threads": "1" } },
  "base": { "name": "ranga-1.20", "options": { "Hash": "128", "Threads": "1" } },

  "tc": "1+0.1",
  "nodes": 5000,

  "rounds": 10,
  "games": 2,
  "concurrency": 4,

  "sprt": { "elo0": 0, "elo1": 2, "alpha": 0.05, "beta": 0.05 },

  "openings": { "file": "books/UHO_Lichess_4852_v1.epd", "format": "epd" },

  "output_dir": "output",
  "extra_args": ["-ratinginterval", "5"]
}
```

The `sprt` block is required, and all four values must be present. `elo0` must be less than `elo1`, and `alpha` and `beta` must be between 0 and 1.

## Selftest

Plays one engine against itself. fastchess won't let two engines share a name, so the tool names them `<name>-a` and `<name>-b`. Both use the same executable and the same options.

```bash
ranga-tools tournament selftest --config selftest.json
```

```json
{
  "engine": { "name": "ranga-1.21", "options": { "Hash": "128", "Threads": "1" } },

  "tc": "1+0.1",
  "nodes": 5000,

  "rounds": 10,
  "games": 1,
  "concurrency": 3,

  "openings": { "file": "books/UHO_Lichess_4852_v1.epd", "format": "epd" },

  "output_dir": "output",
  "extra_args": ["-ratinginterval", "5"]
}
```

## Config reference

These fields work in every tournament type. Only the engine fields (and `sprt`) differ.

| Field | Meaning | Default |
|---|---|---|
| `tc` | Time control, e.g. `"4+0.04"` | none |
| `nodes` | Nodes per move | `5000`, used only when `tc` is not set |
| `rounds` | Number of rounds | `5000` (`50000` for selftest) |
| `games` | Games per pairing in each round | `2` (`1` for selftest) |
| `concurrency` | Games played in parallel | CPU count minus 1, at least 1 |
| `openings.file` | Opening book (`.epd` or `.pgn`) | no book |
| `openings.format` | `epd` or `pgn` | taken from the file extension |
| `openings.order` | `random` or `sequential` | `random` |
| `output_dir` | Folder for results | `output` |
| `extra_args` | Extra arguments passed straight to fastchess | none |

Engine entries (`rc`, `base`, `engine`, `engine1`, `engine2`) have a `name` (a registered engine) and optional `options`, a map of UCI option names to values. Numbers, strings and booleans are all accepted, so `"Hash": 128` and `"Hash": "128"` both work.

**Time control.** If `tc` is set, `nodes` is ignored. If neither is set, each move is limited to 5000 nodes. Use `tc` for timed games or `nodes` for a fixed-node run, not both.

## Output

- **Where.** Results go to `output_dir`, which is relative to the folder containing the config file (not to your current directory). Absolute paths work as written. The same applies to `openings.file`.
- **Files.** `games.pgn` has every game, and `games.log` has the engine communication log.
- **Reruns append.** fastchess adds to existing files, so give each experiment its own folder or `output_dir` to keep results separate.
- **Crashes.** Games continue if an engine crashes (`-recover` is always on).
- **Stopping.** `Ctrl-C` is passed to fastchess so it can stop cleanly and print its summary.
- **Exit status.** A non-zero exit from fastchess is reported as an error with its exit code.

## Validation

The config is checked before fastchess starts, and an invalid config stops the run with a message naming the file:

- Unknown keys are rejected, so a typo like `"round"` fails instead of silently falling back to the default.
- Engine names are required, must be registered, and must be different from each other.
- Negative values for `rounds`, `games`, `concurrency` and `nodes` are rejected.
- Opening `format` and `order` must be one of the listed values.

# Usage Guide for the `analyze` Command

`analyze` runs a registered UCI engine over every game a player appears in within a PGN file, then reports how that player did: accuracy, blunders, record, and their best and worst games.

```bash
ranga-tools analyze --engine <name-of-engine> --pgn <path-to-pgn> --player <name> [options]
```

## Flags

**Required**
- `--engine`: Name of a registered engine to use for the analysis.
- `--pgn`: Path to the PGN file. It may contain many games.
- `--player`: The player whose games are analyzed. This is a case-insensitive regular expression matched against the PGN `White` and `Black` names, so `--player carlsen` matches `Carlsen, Magnus`. If several different names match, the first one found in the file is used, and only games with exactly that name are analyzed. Games where the player faces themselves are skipped.

**Optional**
- `--strength`: Analysis strength: `low`, `medium` or `high`. Default is `low`.
- `--workers`: Number of engines to run in parallel. Default is `1`. Each worker is a separate engine process with its own hash memory, so keep it at or below your CPU core count.
- `--full`: List every game in the report. By default, the report lists only the 5 worst and 5 best games by accuracy (or all games when there are 10 or fewer).
- `--json`: Write the summary to this JSON file.
- `--text`: Write the report to this text file. It has exactly the same content as the terminal output.

If neither `--json` nor `--text` is given, the report is printed to the terminal. If either is given, the report is not printed, and the command only confirms which files were written. You can pass both.

## Examples

```bash
# Print the report to the terminal
ranga-tools analyze --engine ranga-1.23 --pgn games.pgn --player HangMateIn

# Faster, with 6 engines in parallel and deeper analysis
ranga-tools analyze --engine ranga-1.23 --pgn games.pgn --player HangMateIn --strength medium --workers 6

# Save a JSON summary and a text report, listing every game
ranga-tools analyze --engine ranga-1.23 --pgn games.pgn --player HangMateIn --json summary.json --text report.txt --full
```

## What the report contains

- Engine name and strength, number of games analyzed (and failed), and the time taken.
- Overall accuracy, plus accuracy as White and as Black.
- Blunders per game and in total.
- Record, written as `+wins =draws -losses`.
- Best and worst games (or all games with `--full`). Each row shows the game number in the PGN, the opponent, the player's colour (`W` or `B`), the result (`W`, `D`, `L`, or `?` if unknown), the number of plies, accuracy, and a small gauge.

Games that fail to parse or analyze are listed in a "Failed" section and don't stop the run. The command only fails if every game fails.

## JSON output

The JSON file holds the summary only, not per-move data:

- `engine`, `strength`, `depth`, `hash_mb`
- `player`: name, record, and overall, white (`as_white`) and black (`as_black`) aggregates, plus one short entry per game (number, opponent, colour, result, plies, accuracy, class counts)
- `errors`: the failures, present only if there were any

Class counts omit classes with zero moves, so a missing key means 0.

## Notes

- **Strength presets:** `low` searches to depth 8, `medium` to depth 12 and `high` to depth 16, each with 64 MB of hash on a single thread.
- **Accuracy** depends on the engine and strength used, so compare numbers only between runs with the same settings.
- **Standard chess only:** Chess960 and other variants are rejected and listed as failed games. Games starting from a custom FEN are supported.
- **Overwriting:** `--json` and `--text` overwrite existing files without asking.

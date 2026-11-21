# Usage Guide for the `viriformat` Command

The `viriformat` command has utilities for viriformat (`.vf`) training data files: removing duplicate games and shuffling game order.

```bash
ranga-tools viriformat <dedupe|shuffle> [flags]
```

Both commands work on whole games. A game is stored as a 32-byte packed starting position, then a list of 4-byte move and score pairs, then a 4-byte zero terminator. Game bytes are copied unchanged, and nothing is re-encoded.

## Shuffle

Randomizes the order of the games in a file.

```bash
ranga-tools viriformat shuffle --input <path-to-vf> --output <path-to-vf> [--seed <seed>] [--expect-games <n>]
```

**Flags**
- `--input`, `-i`: (Required) The input viriformat file.
- `--output`, `-o`: (Required) Where the shuffled file is written. It must not already exist, since the tool never overwrites.
- `--seed`: (Optional) Seed for the shuffle. The same seed always gives the same order. Default is `0`.
- `--expect-games`: (Optional) Abort if the file doesn't contain exactly this many games. Use it as a sanity check on the file layout.

On success it prints the number of games and bytes. The output is checked to have the same size as the input. If anything fails partway, the partial output file is removed.

**Example**

```bash
ranga-tools viriformat shuffle -i merged.vf -o shuffled.vf --seed 42 --expect-games 1000000
```

## Dedupe

Removes duplicate games, keeping the first copy of each.

```bash
ranga-tools viriformat dedupe --input <path-to-vf> --output <path-to-vf> [--check]
```

**Flags**
- `--input`, `-i`: (Required) An input viriformat file. Repeat the flag to merge several files, for example `-i a.vf -i b.vf`. Duplicates are caught across all inputs.
- `--output`, `-o`: (Required) Where the result is written. It must not be one of the inputs. An existing output file is overwritten.
- `--check`, `-c`: (Optional) After writing, verify the output file's structure: every game is complete, the starting position has at most 32 pieces, and the result byte is valid (0, 1 or 2).

**What counts as a duplicate.** Two games are duplicates when they have the same starting position and the same sequence of moves. The scores and the game result are not compared. If a duplicate has a different result from the copy kept, it is counted as a conflict.

The command prints `kept N, dropped N, result conflicts N`, and with `--check` it adds `ok, N games`.

**Example**

```bash
ranga-tools viriformat dedupe -i a.vf -i b.vf -o merged.vf --check
```

## Memory

- **Dedupe** keeps a 16-byte fingerprint per unique game in memory, so roughly 17 bytes per game.
- **Shuffle** keeps the position of every game in memory (about 20 bytes per game) and reads games back from the input file in random order. Both are small compared with the data itself, but a file with hundreds of millions of games needs a few GB of RAM.

## Typical workflow

```bash
ranga-tools viriformat dedupe -i part1.vf -i part2.vf -o merged.vf --check
ranga-tools viriformat shuffle -i merged.vf -o train.vf --seed 1
```

# Usage Guide for the `engine` Command

The `engine` command manages the UCI chess engines that the other commands (`analyze`, `tournament`) refer to by name. Register an engine once, then use its name everywhere else.

```bash
ranga-tools engine <register|deregister|list|use|bench> [flags]
```

Registered engines are stored in `engines.json` inside the tool's folder in your user config directory (`~/.config/ranga-tools` on Linux, `~/Library/Application Support/ranga-tools` on macOS, `%AppData%\ranga-tools` on Windows).

## Register

Verifies that an executable is a working UCI engine and records it.

```bash
ranga-tools engine register --path <path-to-executable> [--name <name>] [--force]
```

**Flags**
- `--path`: (Required) Path to the engine's executable. It is stored as an absolute path, so it can be relative when you register it.
- `--name`: (Optional) Name to register the engine under. Defaults to the name the engine reports (`id name`). If the engine reports none, `--name` is required.
- `--force`: (Optional) Overwrite an existing engine with the same name. Without it, registering a duplicate is an error.

**What verification does.** The tool starts the engine, sends `uci` and waits for `uciok`, then sends `isready` and waits for `readyok`. Each step has a 5 second timeout. The engine's author (`id author`) is stored too, and is left blank if the engine doesn't send one.

**Example**

```bash
ranga-tools engine register --path ./builds/ranga-1.23 --name ranga-1.23
```

## Names

Names are matched case-insensitively, and spaces and hyphens are interchangeable. `Stockfish 17`, `stockfish-17` and `STOCKFISH 17` all refer to the same engine. This applies to every command that takes an engine name, including the names inside tournament config files.

## Deregister

Removes an engine from the registry. The executable itself is not touched.

```bash
ranga-tools engine deregister --name <name-of-engine>
```

**Flags**
- `--name`: (Required) Name of the engine to remove.

## List

Prints every registered engine, sorted by name, one per line as `name<TAB>author`.

```bash
ranga-tools engine list
```

If nothing is registered, it prints `No engines registered.` to stderr.

## Use

Runs a registered engine with your terminal attached, so you can type UCI commands to it. It behaves the same as running the binary directly.

```bash
ranga-tools engine use --name <name-of-engine>
```

**Flags**
- `--name`: (Required) Name of the engine to run.

Type `quit` to leave. `Ctrl-C` is passed to the engine and not to this tool, so the engine decides how to react. If the engine exits with a non-zero code, the tool reports it and exits with an error. The stored path is used as is, so if the executable has moved, you get an OS error and need to re-register with `--force`.

## Bench

Runs the engine's built-in `bench` command and reports the node count and speed.

```bash
ranga-tools engine bench --name <name-of-engine>
```

**Flags**
- `--name`: The name of the engine to benchmark.

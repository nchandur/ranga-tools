# Usage Guide for the `config` Command

The `config` command reads and writes the tool's settings. Settings are stored in `config.json`, next to `engines.json`, in the tool's folder in your user config directory.

```bash
ranga-tools config set <option> <value>
ranga-tools config get <option>
```

`set` stores a value and confirms with `set <option>`. `get` prints the stored value, or an empty line if the option has never been set. An unknown option name is an error that lists the valid ones.

## Supported Options

### `fastchess-path`

The location of the [fastchess](https://github.com/Disservin/fastchess) binary used by `tournament`.

```bash
ranga-tools config set fastchess-path <path-to-fastchess>
ranga-tools config get fastchess-path
```

The path must point to an existing file. It is stored as an absolute path, so a relative path is fine when you set it.

**How the binary is found.** Tournaments look for fastchess in this order:

1. The `--fastchess <path>` flag on the tournament command (this run only).
2. The stored `fastchess-path` setting.
3. A `fastchess` executable on your `PATH`.

So you only need this setting if fastchess is not on your `PATH`, or if you want to use a specific build. If none of the three is found, the tournament stops with an error that says so.

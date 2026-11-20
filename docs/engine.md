# Usage Guide for the `engine` Command

This guide outlines the various subcommands supported by the `engine` command.

## Register

Before using this tool, you must register your engine to verify it as a valid UCI engine. Use the following command:

```bash
ranga-tools engine register --name <name-of-engine> --path <path-to-executable>
```

**Flags**
- `--name`: (Optional) Assign a custom name to the engine (defaults to whatever it finds in the engine signature)
- `--path`: (Required) The file path to the engine's executable

## Deregister

To remove a registered engine from the tool, use the following command:

```bash
ranga-tools engine deregister --name <name-of-engine>
```

**Flags**
- `--name`: (Required) The name of the engine you want to remove

## List

To view a list of all currently registered engines, use the following command:

```bash
ranga-tools engine list
```

## Use

To execute a registered UCI engine, use the following command. This functions identically to running the binary directly:

```bash
ranga-tools engine use --name <name-of-engine>
```

- `--name`: The name of the engine you want to execute

## Bench

To run bench a registered UCI engine, use the following command.

```bash
ranga-tools engine bench --name <name-of-engine>
```

- `--name`: The name of the engine you want to run bench
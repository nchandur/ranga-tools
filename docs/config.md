# Usage Guide for the `config` Command

This guide explains how to use the `config` command.

## Set 

To set a configuration option, use the following command:
```bash
ranga-tools config set option value
```

## Get
To retrieve the value of a specific configuration option, use the following command:
```bash
ranga-tools config get option
```

## Supported Options

Currently, the config command supports the following options.

### Fastchess Configuration
You must have fastchess installed on your system for tournaments to run.

To set the path to your fastchess installation, use this command:
```bash
ranga-tools config set fastchess-path <path-to-fastchess>
```
> Note: If this path is not set, tournaments will not execute.

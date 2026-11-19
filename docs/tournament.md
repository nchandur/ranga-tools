# Usage Guide for `tournament` command

This guide explains how to use the `tournament` command.

## Prerequisites

Before proceeding, please review the configuration guide [here](docs/config.json)

## Match

This command executes a match-up between two registered engines. Use the following command to run the test: 
```bash
ranga-tools tournament match --config <path-to-config>
```

Both engines involved must be registered with the tool prior to running this command. You can define the match parameters in a JSON configuration file, as shown in the example below:
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

## SPRT

This command executes a Sequential Probability Ratio Test (SPRT) between two registered engines. Use the following command to run the test:

```bash
ranga-tools tournament sprt --config <path-to-config>
```

Both engines involved must be registered with the tool prior to running this command. You can define the SPRT parameters in a JSON configuration file, as shown in the example below:

```json
{
  "rc": { "name": "ranga-1.21", "options": { "Hash": "128", "Threads": "1" } },
  "base": { "name": "ranga-1.20", "options": { "Hash": "128", "Threads": "1" } },

  "tc": "1+0.1",
  "nodes": 5000,

  "rounds": 10,
  "games": 2,
  "concurrency": 4,

  "sprt": { "elo0": 0, "elo1": 2, "alpha": 0.05, "beta": 0.05 },

  "openings": {
    "file": "books/UHO_Lichess_4852_v1.epd",
    "format": "epd"
  },

  "output_dir": "output",
  "extra_args": ["-ratinginterval", "5"]
}
```

The test results will be saved in an output directory located in the same directory as your configuration file.

## Selftest

This command executes a selftest between an engine and itself. Use the following command to run the test:
```bash
ranga-tools tournament selftest --config <path-to-config>
```

The engine must be registered with the tool prior to running this command. You can define the selftest parameters in a JSON configuration file, as shown in the example below:

```json
{
  "engine": {
    "name": "ranga-1.21",
    "options": { "Hash": "128", "Threads": "1" }
  },

  "tc": "1+0.1",
  "nodes": 5000,

  "rounds": 10,
  "games": 1,
  "concurrency": 3,

  "openings": {
    "file": "books/UHO_Lichess_4852_v1.epd",
    "format": "epd"
  },
  
  "output_dir": "output",
  "extra_args": ["-ratinginterval", "5"]
}
```

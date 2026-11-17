# Usage Guide for the viriformat Command

This guide explains how to use the viriformat command.

## Shuffle

This command randomizes the order of games within a viriformat file. Use the following command to shuffle your games:
```bash
ranga-tools viriformat --input <path-to-vf> --output <path-to-vf> --seed <seed>
```

**Flags**
- `--input`: (Required) The file path to the input viriformat file.   
- `--output`: (Required) The file path where the shuffled viriformat file will be saved.   
- `--seed`: (Optional) A seed value to ensure the shuffle order is reproducible.   

## Dedupe

This command deduplicates positions within a viriformat file. Use the following command to run the deduplication process:
```bash
ranga-tools viriformat --input <path-to-vf> --output <path-to-vf> --check
```

**Flags**
- `--input`: (Required) The file path to the input viriformat file.  
- `--output`: (Required) The file path where the deduplicated viriformat file will be saved.   
- `--check`: (Optional) Executes a sanity check on the generated output file.

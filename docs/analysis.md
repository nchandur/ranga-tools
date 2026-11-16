# Usage Guide for the analyze Command

This guide explains how to perform game analysis using the `analyze` command.

This command provides a basic analysis of a chess game supplied in PGN (Portable Game Notation) format. To run the analysis, use the following command:

```bash
ranga-tools analyze --engine <name-of-engine> --pgn <path-to-pgn> --strength <low|medium|high> --json <path-to-json>
```

**Flags**
- `--engine`: (Required) The name of a registered engine to use for the analysis.
- `--pgn`: (Required) The file path to the game PGN. If the file contains multiple games, the tool will only analyze the first one   
- `--strength`: (Optional) The depth or strength of the analysis. You can specify low, medium, or high. The default value is low.   
- `--json`: (Required) The file path where the final analysis report will be saved in JSON format.

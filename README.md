# Halstead Metrics for Go

`halstead-metrics` is a Go project for measuring and controlling complexity in Go code, especially complexity introduced by AI-assisted code generation.

The analyzer is built on top of Go's standard tooling:

- `go/parser` for parsing source files
- `go/ast` for traversing syntax trees
- `go/types` for semantic symbol resolution

That makes the output much closer to what the Go compiler understands than a plain token-based parser.

The project is intended to help teams answer questions like:

- Did this AI-generated change make the file or function materially more complex?
- Which function absorbed the complexity increase?
- Should this change be accepted, revised, or split into smaller pieces?
- Can we enforce complexity limits automatically in editors, CI, or code review?

## Why This Matters for AI-Generated Code

AI can generate useful code very quickly, but it can also introduce complexity just as quickly:

- overly dense helper functions
- large multi-branch flows
- unnecessary abstractions
- hidden complexity moved into "convenient" wrappers
- code that is locally correct but harder for humans to maintain

This project is aimed at making that complexity visible early, before it spreads through a codebase.

A practical workflow looks like this:

1. Generate or edit code with AI assistance.
2. Run `halstead` on the changed file or package.
3. Compare file-level and function-level complexity against team thresholds.
4. Flag suspicious increases during review, in Emacs, or in CI.
5. Use the report to ask for simplification before merging.

## Features

- Analyzes real Go syntax using the standard library parser
- Resolves symbols semantically with `go/types`
- Reports operators and operands separately
- Reports file-level and per-function metrics
- Helps identify complexity added by generated code at the function level
- Distinguishes semantic operands such as `var:x`, `func:Println`, `type:string`, `pkg:fmt`
- Computes standard Halstead-derived values such as vocabulary, length, volume, difficulty, effort, time, and estimated bugs
- Includes 20 example Go programs in [`testdata/`](testdata/)

## Project Goal

The goal is not just to compute metrics.

The goal is to make complexity visible enough that teams can use it as a guardrail when AI is writing or reshaping production code.

In that framing, `halstead-metrics` is best used as:

- a review aid for AI-generated pull requests
- an editor-side warning system for complexity spikes
- a CI quality gate for generated or heavily transformed code
- a source of machine-readable complexity data for custom tooling

## Counting Policy

Current counting rules:

- Counts executable and declaration-related constructs as operators, such as `func`, `var`, `const`, `type`, `if`, `for`, `switch`, `return`, assignments, calls, selectors, indexing, slicing, and type assertions
- Does not count `package` or `import` as operators, because they describe file organization and dependencies rather than program behavior
- Counts semantically resolved symbols as operands, such as `var:x`, `func:Println`, `builtin:make`, `type:string`, `pkg:fmt`, and `field:name`
- Counts literals as operands, such as `"hello"` or `42`

## Requirements

- Go 1.22 or newer

## Installation

Clone the repository and build the CLI:

```bash
go build -o halstead ./cmd/halstead
```

If you want to install the binary from a local checkout, use:

```bash
go install ./cmd/halstead
```

Why not `go install .`?

- `go install .` targets the package in the repository root
- the repository root is a library package, not the CLI entry point
- the executable lives in [`cmd/halstead`](cmd/halstead)

If the project is published at its module path, you can also install it remotely with:

```bash
go install github.com/luisantonioig/halstead-metrics/cmd/halstead@latest
```

## Usage

Analyze a Go source file:

```bash
./halstead testdata/ejem_01.go
```

Emit machine-readable JSON:

```bash
./halstead --json testdata/ejem_01.go
```

Fail when configured thresholds are exceeded:

```bash
./halstead --max-volume 20 --max-difficulty 3 testdata/ejem_15.go
```

Compare the current file against a saved baseline report:

```bash
./halstead --baseline-report baseline.json testdata/ejem_15.go
```

Compare the current file against its version at a Git revision:

```bash
./halstead --baseline-git HEAD~1 path/to/file.go
```

Focus the comparison on changed or newly added functions only:

```bash
./halstead --baseline-report baseline.json --changed-only testdata/ejem_15.go
```

Fail when complexity growth exceeds a configured delta budget:

```bash
./halstead --baseline-report baseline.json --max-volume-delta 15 testdata/ejem_15.go
```

The output includes:

- operators found
- operands found
- semantically classified operands
- per-function summaries with line/column locations
- number of distinct operators and operands
- program length
- volume
- difficulty
- effort
- estimated time to program
- estimated delivered bugs

## Quick Start

From the repository root:

```bash
go test ./...
go build -o halstead ./cmd/halstead
./halstead testdata/ejem_01.go
./halstead --json testdata/ejem_01.go
```

## JSON Output

`--json` is intended for editor, CI, and automation integrations.

The JSON report includes:

- `analyzer`
- `path`
- `file`: aggregate metrics for the whole file
- `functions`: per-function metrics with name, kind, source range, and Halstead values
- `comparison`: optional baseline deltas for the file and each function

This makes it suitable for:

- Emacs integrations through `flymake`, `flycheck`, or custom commands
- CI quality gates
- scripts that compare complexity before and after changes
- workflows that review AI-generated patches before merge
- future LSP-style diagnostics or code lenses

## Baseline Comparison

To make AI-generated complexity increases visible, you can compare the current file against a previously saved JSON report.

Typical flow:

1. Save a baseline report from trusted code:

```bash
./halstead --json path/to/file.go > baseline.json
```

2. Analyze the updated or AI-generated version against that baseline:

```bash
./halstead --baseline-report baseline.json path/to/file.go
```

3. Enforce a complexity growth budget:

```bash
./halstead \
  --baseline-report baseline.json \
  --max-volume-delta 15 \
  --max-difficulty-delta 2 \
  path/to/file.go
```

The comparison output shows:

- file-level deltas
- per-function deltas
- whether a function is new in the current report
- threshold failures for complexity growth

If you add `--changed-only`, the output and JSON focus only on functions whose complexity changed relative to the baseline, plus newly added functions.

If you already have the baseline in Git, you can skip the saved JSON step and compare directly against a revision:

```bash
./halstead \
  --baseline-git HEAD~1 \
  --changed-only \
  --max-volume-delta 15 \
  path/to/file.go
```

## Emacs Integration

There are two good ways to use `halstead` in Emacs without building a full language server.

### Option 1: Run it on demand

This is the simplest workflow. Add a small helper command to your Emacs config and inspect the JSON output in a buffer:

```elisp
(defun halstead-analyze-current-file ()
  "Run halstead on the current Go buffer and show JSON output."
  (interactive)
  (unless buffer-file-name
    (user-error "Current buffer is not visiting a file"))
  (let ((buf (get-buffer-create "*halstead*")))
    (with-current-buffer buf
      (erase-buffer))
    (call-process "halstead" nil buf nil "--json" buffer-file-name)
    (display-buffer buf)))
```

This is a good first step if you want to inspect file and function metrics manually.

### Option 2: Use Flycheck as a quality gate

Because the CLI now supports thresholds and exits with `1` when they fail, it can be used as a lightweight checker.

Example `flycheck` configuration:

```elisp
(with-eval-after-load 'flycheck
  (flycheck-define-checker go-halstead
    "Run halstead thresholds on the current Go file."
    :command ("halstead"
              "--max-volume" "80"
              "--max-difficulty" "8"
              source)
    :error-patterns
    ((warning line-start "Thresholds" (zero-or-more anything) line-end))
    :modes go-mode)

  (add-to-list 'flycheck-checkers 'go-halstead))
```

This example is intentionally minimal. In practice, the best setup is usually:

- keep `halstead --json` as the machine-readable API
- use a small Emacs wrapper to parse JSON
- surface warnings per function using the reported line and column ranges

### Recommended editor roadmap

If you want this to become a stronger development-cycle tool, a good progression is:

1. Use `--json` from Emacs commands to explore the data model.
2. Add threshold-based checks for save-time feedback.
3. Parse function-level JSON into `flymake` or `flycheck` diagnostics.
4. Optionally build a small LSP-style server later if you want richer UI features such as code lenses or hover summaries.

## Thresholds and Exit Codes

The CLI can enforce simple quality gates:

- `--max-volume`
- `--max-difficulty`
- `--max-effort`
- `--max-volume-delta`
- `--max-difficulty-delta`
- `--max-effort-delta`
- `--changed-only`
- `--baseline-git`

Thresholds are evaluated against:

- the whole file
- each individual function report
- optional baseline deltas when `--baseline-report` is used

Exit codes:

- `0`: analysis succeeded and all configured thresholds passed
- `1`: analysis failed or at least one configured threshold failed
- `2`: invalid CLI usage

Example:

```bash
./halstead --max-volume 40 --max-difficulty 4 testdata/ejem_15.go
./halstead --json --max-effort 100 testdata/ejem_18.go
```

For AI-assisted development, a useful pattern is:

```bash
./halstead --json --max-volume 80 --max-difficulty 8 path/to/generated_file.go
```

If the command exits with `1`, the change exceeded your complexity budget and should be reviewed or simplified before merge.

For baseline-aware reviews of generated code:

```bash
./halstead \
  --json \
  --baseline-git origin/main \
  --changed-only \
  --max-volume-delta 20 \
  --max-difficulty-delta 3 \
  path/to/generated_file.go
```

If this exits with `1`, the generated change increased complexity more than your configured budget.

## CI and Automation

`halstead` is designed to work well in automation because it has machine-readable JSON output and stable exit codes.

### GitHub Actions example

This example builds the CLI and fails the job if a generated or edited file exceeds the configured complexity budget:

```yaml
name: halstead

on:
  pull_request:
  push:
    branches: [main]

jobs:
  complexity:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Build halstead
        run: go build -o ./bin/halstead ./cmd/halstead

      - name: Check complexity budget
        run: |
          ./bin/halstead \
            --baseline-git origin/main \
            --changed-only \
            --max-volume-delta 20 \
            --max-difficulty-delta 3 \
            path/to/generated_file.go
```

If your CI checkout does not already have the target branch available locally, fetch it before running the comparison:

```bash
git fetch origin main
```

### Pre-commit or local hook example

For a simple local quality gate, run `halstead` against a file before committing:

```bash
./halstead \
  --max-volume 80 \
  --max-difficulty 8 \
  path/to/file.go
```

For AI-assisted changes, a baseline-aware command is usually more useful:

```bash
./halstead \
  --baseline-git HEAD~1 \
  --changed-only \
  --max-volume-delta 20 \
  --max-difficulty-delta 3 \
  path/to/file.go
```

## Example Programs

The [`testdata/`](testdata/) directory contains 20 Go examples:

- `ejem_01.go` to `ejem_05.go`: simple functions, imports, and calls
- `ejem_06.go` to `ejem_10.go`: parameters, multiple returns, variables, and expressions
- `ejem_11.go`: `if` and unary operators
- `ejem_12.go`: slices and `for range`
- `ejem_13.go`: maps and indexing
- `ejem_14.go`: structs and composite literals
- `ejem_15.go`: methods with receivers
- `ejem_16.go`: `switch`, `case`, and `default`
- `ejem_17.go`: `defer`
- `ejem_18.go`: goroutines and channels
- `ejem_19.go`: `select` with channel receive
- `ejem_20.go`: type switch and type assertion

You can explore them manually:

```bash
./halstead testdata/ejem_14.go
./halstead testdata/ejem_19.go
```

## Development

Run tests:

```bash
go test ./...
```

Build the CLI:

```bash
go build ./cmd/halstead
```

Format Go code:

```bash
gofmt -w *.go cmd/halstead/*.go
```

## Project Structure

- [`cmd/halstead`](cmd/halstead): CLI entry point
- [`testdata/`](testdata/): example Go files used for testing and exploration
- repository root package: analysis logic and Halstead calculations

## Contributing

Issues and pull requests are welcome. If you change the counting policy, please update tests and documentation so behavior stays explicit.

Good contribution directions for the AI-complexity use case include:

- package-level and repository-level reports
- diff-aware analysis for changed functions only
- baseline comparison against the target branch
- CI examples for pull request gating
- editor integrations that turn function reports into diagnostics
- better heuristics for spotting suspicious complexity jumps in generated code

## License

No license file is included yet. Before publishing broadly as open source, add a `LICENSE` file with the terms you want to use.

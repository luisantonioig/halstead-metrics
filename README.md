# Halstead Metrics for Go

Catch complexity growth in Go code before merging AI-assisted changes.

[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](go.mod)

## 30-second example

After [installing](#install), paste this into a POSIX shell. It creates an isolated
Go module and compares a simple function with an unnecessarily complex rewrite:

```bash
cd "$(mktemp -d)"
go mod init example.com/halstead-demo
cat > price.go <<'GO'
package demo

func total(price, quantity int) int {
    return price * quantity
}
GO
halstead --json price.go > baseline.json
cat > price.go <<'GO'
package demo

func total(price, quantity int) int {
    result := price * quantity
    if quantity == 0 { return 0 }
    if quantity == 1 { return price }
    return result
}
GO
halstead --baseline-report baseline.json --changed-only --max-volume-delta 20 price.go
```

Actual output excerpt (colors omitted); the last command exits with **1**:

```text
File delta:        volume 64.44, difficulty 3.80, effort 521.02
Volume:            FAILED delta 64.44, limit 20.00
Functions:         1 compared, 1 changed or new
Status:            FAILED 1 function delta(s) exceeded the configured thresholds
```

Both versions compute the same result. Volume grows from **28.07 to 92.51**;
the redundant branches exceed the chosen growth budget. Add `--verbose` to see
which function exceeded it. This is an illustrative rewrite, not an AI benchmark.

## Install

Requires Go 1.22+ on `PATH`, including when running the analyzer.

```bash
go install github.com/luisantonioig/halstead-metrics/cmd/halstead@latest
```

Put Go's install directory (`go env GOPATH` followed by `/bin`, or your custom
`GOBIN`) on `PATH`. Then, in the demo directory above:

```bash
halstead price.go
```

For your project, replace `price.go` with one existing `.go` file in a buildable
Go package. Run once per file; package paths such as `./...` are not CLI inputs.

## Why?

AI-assisted edits can preserve behavior while adding branches, repeated work,
and unnecessary abstractions. `halstead` makes changes in Halstead volume,
difficulty, and effort visible at file and function level, so reviewers can ask
for simplification before merging.

These metrics are a review signal and a configurable guardrail, not a verdict
on code quality or a detector of AI authorship. A larger value can be justified;
choose budgets for your codebase and review the underlying change.

## CI / Pull Request guardrail

From this repository's root, with `origin/master` available locally:

```bash
halstead \
  --baseline-git origin/master \
  --changed-only \
  --max-volume-delta 20 \
  --max-difficulty-delta 3 \
  metrics.go
```

Use your target branch and source file in your own repository. Growth beyond
either budget in the file or a compared function exits **1**; passing exits **0**.
Analysis errors also exit **1**, so inspect the report or stderr before concluding
that a budget failed. Deltas are absolute metric units, not percentages.

In GitHub Actions, use `actions/checkout@v7` with `fetch-depth: 0` so the base
branch exists, set up Go with `actions/setup-go@v7`, install the CLI, and run the
command above as a step. For a PR to any branch, pass its base SHA via an environment
variable from `github.event.pull_request.base.sha` instead of hard-coding a branch.
The [project CI](.github/workflows/ci.yml) runs tests, vet, and the CLI build.
The same comparison command can be used in a local pre-commit hook.

## Features

- File and function metrics, including methods and function literals.
- AST analysis with semantic symbol resolution through `go/types`.
- Terminal summaries, detailed counts, and JSON for editor/CI integrations.
- Absolute thresholds and growth budgets for volume, difficulty, and effort.
- Saved JSON or Git revision baselines; `--changed-only` filters function reports.
- Standard library only; no third-party Go dependencies.

## Detailed usage

The following commands use `price.go` and `baseline.json` from the demo:

```bash
halstead --verbose price.go
halstead --json price.go
halstead --max-volume 80 --max-difficulty 8 price.go
halstead --json --max-effort 100 price.go
halstead --baseline-report baseline.json price.go
halstead --baseline-report baseline.json --changed-only --max-volume-delta 15 price.go
```

The threshold examples deliberately exit **1**. Without thresholds, analysis
reports values without enforcing a limit.

### Thresholds and exit codes

| Flags | Meaning |
| --- | --- |
| `--max-volume`, `--max-difficulty`, `--max-effort` | Limit absolute file and function metrics. |
| `--max-volume-delta`, `--max-difficulty-delta`, `--max-effort-delta` | Limit growth; requires a baseline to have an effect. |
| `--baseline-report report.json` | Compare against saved JSON from `--json`. |
| `--baseline-git REV` | Compare against that revision's version of the same file. |
| `--changed-only` | With a baseline, retain functions with changed metrics or newly added functions. |

A positive limit enables a check; zero or a negative value disables it.
Use only one baseline flag. Save a JSON baseline before editing the source,
as in the demo. Git baselines can also use `HEAD~1` or a commit SHA.

- **0:** analysis succeeded and configured checks passed.
- **1:** analysis failed or a threshold was exceeded.
- **2:** invalid flag usage or missing input; help also exits 2.

`--changed-only` compares metric values, not changed lines. File metrics and file
budgets remain active. Functions match by name and kind; new functions compare
against zero, removed functions are not listed, and anonymous functions use
source positions as names.

### JSON and terminal output

JSON contains `analyzer`, `path`, `file`, and `functions` (name, kind, source range,
and metrics), plus optional `threshold` and `comparison` outcomes. Comparisons
include baseline/current values, deltas, new-function markers, and violations.
JSON is also emitted when a budget fails, making it useful for CI scripts,
Flymake/Flycheck wrappers, or other editor diagnostics.

Terminal output shows a file summary and the functions with the highest volume
and difficulty. `--verbose` adds operator/operand counts, function locations,
per-function metrics, and comparison details.

### Build from a local checkout

From the repository root:

```bash
go build -o halstead ./cmd/halstead
go install ./cmd/halstead
```

The root package is a library; the install target is `./cmd/halstead`.

### Package context and Git baselines

The CLI uses `go list` to resolve the target package and dependencies. The package
must build in your current Go environment; standalone files outside a module,
excluded files, and directories containing conflicting example programs may fail.

A Git baseline must contain the target file. Its source is analyzed using the
**current** surrounding package and dependencies, not a full historical checkout.
If related declarations changed incompatibly, use a JSON report captured before
the edit. Use the same analyzer version for both reports.

### Emacs integration

Run the analyzer on a saved file and inspect its JSON in a buffer:

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

For inline diagnostics, parse JSON and map function ranges and violations to
Flymake or Flycheck. The terminal `Thresholds` heading appears for passing checks
too, so matching that heading alone does not identify a violation.

## How it works

`go/parser` builds the AST, `go/ast` traverses it, and `go/types` resolves symbols.
The root package exposes `AnalyzeAST`, `AnalyzeASTReport`, `AnalyzeASTFile`, and
`AnalyzeASTSource`; the CLI lives in [cmd/halstead](cmd/halstead).

### Counting policy

Current counting rules:

- Counts executable and declaration-related constructs as operators, such as `func`, `var`, `const`, `type`, `if`, `for`, `switch`, `return`, assignments, calls, selectors, indexing, slicing, and type assertions
- Does not count `package` or `import` as operators, because they describe file organization and dependencies rather than program behavior
- Counts semantically resolved symbols as operands, such as `var:x`, `func:Println`, `builtin:make`, `type:string`, `pkg:fmt`, and `field:name`
- Counts literals as operands, such as `"hello"` or `42`

### Derived metrics

Vocabulary is the number of distinct operators plus operands; length is their
total occurrence count. Volume is length × log₂(vocabulary), difficulty is
(distinct operators / 2) × (total operands / distinct operands), and effort is
difficulty × volume. Zero denominators are handled explicitly.

Reports also include calculated length and traditional Halstead time/bug
estimates. Those estimates are formula outputs, not measured development time
or observed defects; do not use them as forecasts.

### Example programs

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

These are independent library test fixtures, not a single buildable package.
To inspect one with the CLI, from the repository root copy it into its own module:

```bash
example_dir="$(mktemp -d)"
cp testdata/ejem_14.go "$example_dir/main.go"
cd "$example_dir"
go mod init example.com/halstead-fixture
GOFLAGS=-buildvcs=false halstead main.go
```

This disables VCS stamping for the temporary example module.

## Contributing

Found an unexpected count or an analysis error? Open an
[issue](https://github.com/luisantonioig/halstead-metrics/issues) with a small Go
example, the command, and expected/actual output. See [CONTRIBUTING.md](CONTRIBUTING.md)
for local checks and pull requests.

## License

No license has been selected or included yet. Explicit reuse terms are pending
the owner's decision.

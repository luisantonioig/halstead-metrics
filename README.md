# Halstead Metrics for Go

`halstead-metrics` is a small Go project that computes Halstead metrics from Go source code.

The analyzer is built on top of Go's standard tooling:

- `go/parser` for parsing source files
- `go/ast` for traversing syntax trees
- `go/types` for semantic symbol resolution

That makes the output much closer to what the Go compiler understands than a plain token-based parser.

## Features

- Analyzes real Go syntax using the standard library parser
- Resolves symbols semantically with `go/types`
- Reports operators and operands separately
- Distinguishes semantic operands such as `var:x`, `func:Println`, `type:string`, `pkg:fmt`
- Computes standard Halstead-derived values such as vocabulary, length, volume, difficulty, effort, time, and estimated bugs
- Includes 20 example Go programs in [`testdata/`](/home/antonio/personal/halstead-metrics/testdata)

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

## Usage

Analyze a Go source file:

```bash
./halstead testdata/ejem_01.go
```

The output includes:

- operators found
- operands found
- semantically classified operands
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
```

## Example Programs

The [`testdata/`](/home/antonio/personal/halstead-metrics/testdata) directory contains 20 Go examples:

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

- [`cmd/halstead`](/home/antonio/personal/halstead-metrics/cmd/halstead): CLI entry point
- [`testdata/`](/home/antonio/personal/halstead-metrics/testdata): example Go files used for testing and exploration
- repository root package: analysis logic and Halstead calculations

## Contributing

Issues and pull requests are welcome. If you change the counting policy, please update tests and documentation so behavior stays explicit.

## License

No license file is included yet. Before publishing broadly as open source, add a `LICENSE` file with the terms you want to use.

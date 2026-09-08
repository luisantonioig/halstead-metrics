# Contributing

For bugs, include a minimal Go example, your Go version, the exact command,
and expected versus actual output. Mention build tags or dependencies if relevant.

Keep pull requests focused. For behavior or counting changes, add a regression
test and update the documentation so the counting policy stays explicit.

From the repository root, with Go 1.22+ installed:

```bash
gofmt -w .
go test ./...
go vet ./...
go build ./cmd/halstead
git diff --check
```

The root is the library; the executable entry point is `cmd/halstead`.
The independent programs in `testdata` are analyzed by library tests.

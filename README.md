# Runvil Libraries

**Runvil Libraries** is a monorepo of modular, reusable libraries written in [Go](https://go.dev/) for the Runvil ecosystem. It hosts the shared building blocks that power the [Runvil meta-framework](https://github.com/runvil/runvil-framework) and any application built on top of it.

## Overview

Organized as a Go module monorepo, `runvil-libs` exposes each concern as an independent package. Libraries are versioned and released on their own schedule, so consumers pull in exactly what they need without inheriting unused dependencies. A shared core package defines the common primitives that keep the ecosystem consistent.

Wherever the Go standard library already provides the functionality, Runvil intentionally does not re-implement it — the ecosystem standardizes on `flag`, `log/slog`, and `os` rather than shipping parallel packages.

## Features

- **Modular package layout** — One focused package per concern; compose as needed.
- **Stdlib-first** — Standard library packages are used instead of re-implemented.
- **Safe by design** — Go's memory safety with no manual memory management; `unsafe` is not used.
- **Shared conventions** — Consistent formatting and vetting across all packages.

## Workspace Layout

| Path        | Description                                                    |
| ----------- | -------------------------------------------------------------- |
| `core/`     | Shared primitives: common error type and exit-code mapping.    |
| `term/`     | Terminal I/O, output formatting, and color conventions.        |

The Go standard library covers the remaining CLI concerns directly:

| Concern         | Stdlib package                     |
| --------------- | ---------------------------------- |
| Argument parsing | [`flag`](https://pkg.go.dev/flag)  |
| Structured logging | [`log/slog`](https://pkg.go.dev/log/slog) |
| Configuration   | [`os`](https://pkg.go.dev/os) environment variables and [`strconv`](https://pkg.go.dev/strconv) |

## Getting Started

### Prerequisites

- Go toolchain 1.22 or newer — see [go.dev/dl](https://go.dev/dl/)

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

## Libraries

- `core` — Shared primitives, including the common error type and exit-code mapping.
- `term` — Terminal I/O, output formatting, and color conventions.

The following libraries are planned:

- `runvil-http-client` — HTTP client abstractions.

## Contributing

Contributions are welcome. Please run `gofmt` and `go vet` before submitting changes, and ensure all tests pass.

## License

Runvil Libraries is distributed under the [MIT License](LICENSE).

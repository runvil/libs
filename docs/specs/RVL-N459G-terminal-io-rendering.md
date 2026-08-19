# Specification — Terminal I/O & Rendering

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | RVL-N459G                                   |
| Title       | Terminal I/O & Rendering                    |
| Status      | Draft                                       |
| Date        | 2026-08-18                                  |
| Author      | Runvil Contributors                         |
| Domain      | Libraries / SDK — Terminal                  |

## 1. Context

The `term` package provides the terminal-facing primitives consumed across the
Runvil ecosystem: ANSI color and style conventions, terminal capability
detection, and safe output rendering. It is the foundation for the framework's
output and help features.

This specification formalizes the behavior and conventions of `term`.

## 2. Problem Statement

Go CLIs render terminal output inconsistently: hard-coded ANSI escapes leak
into logs and piped output, `NO_COLOR`/`TERM=dumb` conventions are ignored, and
styling vocabulary differs across tools. Output is non-portable and hostile to
machine consumption.

## 3. Goals

- G1 — Provide a consistent, minimal ANSI color and style vocabulary.
- G2 — Guard against emitting escape sequences when the destination cannot render them.
- G3 — Keep the package framework-agnostic and pure-Go (no external dependencies).
- G4 — Remain forward-compatible with streaming output and progress primitives.

## 4. Non-Goals

The following are explicitly out of scope for the initial phase:

- NG1 — Full terminal emulation, raw-mode input handling, or line editors.
- NG2 — A comprehensive progress-bar/spinner library; only the rendering primitives are covered.
- NG3 — Platform-specific terminal APIs (e.g. Win32 console internals) beyond ANSI emission.

## 5. Requirements

### 5.1 Color & Style

| ID         | Requirement                                                       | Priority |
| ---------- | ----------------------------------------------------------------- | -------- |
| RVL-TM-001 | Expose the ANSI 16-color foreground palette (black..white, default). | Must   |
| RVL-TM-002 | Expose at least `bold`, `dim`, and `underline` styles.            | Must     |
| RVL-TM-003 | Provide a paint function that emits `\x1b[<codes>m` + reset per call. | Must  |

### 5.2 Capability Detection

| ID         | Requirement                                                       | Priority |
| ---------- | ----------------------------------------------------------------- | -------- |
| RVL-TM-004 | Detect color support from `NO_COLOR` and `TERM=dumb` conventions. | Must     |
| RVL-TM-005 | Provide a terminal object whose paint degrades to plain text when color is unsupported. | Must |

### 5.3 Output Conventions

| ID         | Requirement                                                       | Priority |
| ---------- | ----------------------------------------------------------------- | -------- |
| RVL-TM-006 | Distinguish data output (stdout) from diagnostics (stderr).       | Must     |
| RVL-TM-007 | Keep escape emissions data-stable: identical input yields identical bytes. | Must |

## 6. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Performance.** Rendering must not allocate beyond the produced string.
- NFR3 — **Portability.** Linux, macOS, and Windows.
- NFR4 — **Quality gates.** `gofmt`, `go vet ./...`, and `go test ./...` must pass.

## 7. Success Criteria

- S1 — Paint output matches the ANSI convention exactly and is covered by table-driven tests.
- S2 — A degraded-terminal configuration emits no escape sequences.
- S3 — The framework's help and output features render solely through `term`.

## 8. References

- [RVL-4Y8UP](./RVL-4Y8UP-runvil-libraries.md) — Runvil Libraries initial specification.
- [RVL-CHBZ4](./RVL-CHBZ4-errors-exit-codes.md) — Core Errors & Exit Codes.
- [RVF-WXQQ5](https://github.com/runvil/framework/blob/main/docs/specs/RVF-WXQQ5-cli-output-formatting.md) — CLI Output & Formatting.
- [RVF-LJWEB](https://github.com/runvil/framework/blob/main/docs/specs/RVF-LJWEB-cli-help-usage.md) — CLI Help & Usage.
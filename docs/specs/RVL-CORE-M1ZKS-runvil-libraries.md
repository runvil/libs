# Specification — Runvil Libraries

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | RVL-M1ZKS                                   |
| Title       | Runvil Libraries — Initial Specification    |
| Status      | Draft                                       |
| Date        | 2026-08-18                                  |
| Author      | Runvil Contributors                         |
| Domain      | Libraries / SDK                             |

## 1. Context

Runvil Libraries (`libs`) is a monorepo of modular, reusable libraries
written in Go for the **Runvil ecosystem**. It hosts the shared building
blocks that power the [Runvil Framework](https://github.com/runvil/framework)
and any application built on top of it.

Unlike the framework, these libraries are **ecosystem-agnostic**: they are
standalone, individually versioned, and usable without the framework. The
framework composes them rather than re-implementing them.

Wherever the Go standard library already provides the functionality, the
ecosystem standardizes on it instead of shipping parallel packages.

### 1.1 Current State

- Module root `github.com/runvil/libs` with `core/` and `term/` packages.
- `core/` provides the shared primitives, including the common error type and exit-code mapping.
- `term/` provides terminal I/O, output formatting, and color conventions.
- Shared module root: Go 1.22, MIT license.

## 2. Problem Statement

The Runvil ecosystem lacks a canonical home for shared, reusable primitives.
Every consumer re-implements terminal I/O, error/exit-code handling, and output
conventions inline or via unrelated third-party packages, fragmenting behavior
and coupling libraries to a framework's assumptions.

## 3. Goals

- G1 — Provide modular, independently versioned libraries that serve the entire Runvil ecosystem.
- G2 — Deliver the foundational CLI building blocks required by the framework's initial (CLI) phase.
- G3 — Keep libraries framework-agnostic and reusable across ecosystems.
- G4 — Maintain production quality: memory safety, documented public APIs, and CI-enforced formatting/vetting.

## 4. Non-Goals

The following are explicitly out of scope for the initial phase:

- NG1 — Embedding framework-level logic or application orchestration in these libraries.
- NG2 — Re-implementing functionality already provided by the Go standard library (`flag`, `log/slog`, `os`, `strconv`).
- NG3 — Guaranteeing a single-version-for-all cadence; libraries version independently.

## 5. Scope — Initial Phase: CLI Building Blocks

To support the framework's CLI ecosystem, `libs` will provide the
reusable foundation primitives. The framework's `cli` package then composes
these into its integrated application model.

### 5.1 Library Capabilities

| ID          | Requirement                                                            | Priority |
| ----------- | ----------------------------------------------------------------------- | -------- |
| RVL-CLI-001 | Maintain `core` as the shared foundation of primitives.                | Must     |
| RVL-CLI-002 | Use the Go standard library `flag` package for argument parsing (no bespoke parser). | Must |
| RVL-CLI-003 | Use the Go standard library `log/slog` package for structured logging. | Must     |
| RVL-CLI-004 | Use the Go standard library `os`/`strconv` for environment-based configuration. | Should |
| RVL-CLI-005 | Provide a `term` package for terminal I/O, output formatting, and color conventions. | Should |
| RVL-CLI-006 | Define a common error type and exit-code mapping usable by all CLI-facing packages. | Must |

### 5.2 Deliverables

- D1 — `core/` extended with the common error type and exit-code mapping (RVL-CLI-006).
- D2 — `term/` implementing RVL-CLI-005.
- D3 — Argument parsing (`flag`), logging (`log/slog`), and configuration (`os`/`strconv`) sourced from the Go standard library (RVL-CLI-002..004).
- D4 — CI workflow enforcing `gofmt`, `go vet`, and `go test ./...`.

## 6. Architecture Constraints

- C1 — The module is a Go monorepo; every package is independently versioned.
- C2 — The module root defines the shared language version and license.
- C3 — Dependencies flow **upward**: higher-level libraries may depend on lower-level libraries, never the reverse; cyclic dependencies are prohibited.
- C4 — Libraries must not depend on framework packages.
- C5 — The `unsafe` package must not be used; all exported identifiers must be documented.
- C6 — New capabilities are introduced as additive packages; breaking changes are not allowed in minor/patch releases.

## 7. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Performance.** Libraries must not impose measurable overhead on dependent applications.
- NFR3 — **Portability.** All packages must target Linux, macOS, and Windows.
- NFR4 — **Minimum Go version.** The minimum supported Go version must be documented in the README.
- NFR5 — **Quality gates.** `gofmt`, `go vet ./...`, and `go test ./...` must pass in CI.

## 8. Success Criteria

- S1 — Every planned package exists, passes `go vet` with no findings, and passes its tests.
- S2 — Documentation comments (rendered by `go doc`) cover every exported identifier.
- S3 — The framework's `cli` package (Phase P1) builds against these libraries.
- S4 — Libraries are usable standalone without depending on the framework.

## 9. Related Specifications

| SpecID    | Title                                                     |
| --------- | --------------------------------------------------------- |
| [RVL-R934Y](./RVL-TERM-R934Y-terminal-io-rendering.md) | Terminal I/O & Rendering                         |
| [RVL-W0J2X](./RVL-CORE-W0J2X-errors-exit-codes.md)     | Core Errors & Exit Codes                          |

## 10. References

- [Runvil Framework — RVF-CMBZJ](https://github.com/runvil/framework/blob/main/docs/specs/RVF-META-CMBZJ-runvil-meta-framework.md) — Initial specification for the Runvil meta-framework.
- [framework](https://github.com/runvil/framework) — the meta-framework consuming these libraries.
- Project `README.md` for building and testing instructions.
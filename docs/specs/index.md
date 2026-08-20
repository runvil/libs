# Specifications Index — Runvil Libraries

This directory holds the formal specifications for the Runvil Libraries.

Ordered by **impact-to-effort** (high impact, low effort first) and **dependency chain** (foundations first).

| SpecID    | Title                                      | Status | Depends On |
| --------- | ------------------------------------------ | ------ | ---------- |
| [RVL-M1ZKS](./RVL-CORE-M1ZKS-runvil-libraries.md) | Runvil Libraries — Initial Specification | Draft | — |
| [RVL-2X1QZ](./RVL-CONFIG-2X1QZ-configuration-loading.md) | Configuration Loading (YAML + env overlay) | Draft | RVL-M1ZKS |
| [RVL-LHANF](./RVL-VALIDATE-LHANF-struct-validation.md) | Struct Validation (required/min/max/len/email/pattern/oneof) | Draft | RVL-M1ZKS |
| [RVL-W0J2X](./RVL-CORE-W0J2X-errors-exit-codes.md) | Core Errors & Exit Codes | Draft | RVL-M1ZKS |
| [RVL-R934Y](./RVL-TERM-R934Y-terminal-io-rendering.md) | Terminal I/O & Rendering | Draft | RVL-M1ZKS |

## Conventions

- Each specification is stored as a single Markdown file named `{ProjectId}-{Scope}-{SpecID}-{slug}.md`.
- SpecIDs are unique, random 5-character alphanumeric codes (e.g., `RVL-M1ZKS`).
- New specifications must be added to this index when created.
- Ordering: impact-to-effort (high impact, low effort first), then dependency chain (foundations first).
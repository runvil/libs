# Specification — Configuration Loading

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | RVL-X7C4M                                   |
| Title       | Configuration Loading                       |
| Status      | Draft                                       |
| Date        | 2026-08-19                                  |
| Author      | Runvil Contributors                         |
| Domain      | Libraries — Configuration                   |

## 1. Context

Runvil applications (CLI tools, servers, static builders) all need typed
configuration resolved from files and the environment. Today each consumer
hand-rolls loading: `runvil` parses `runvil.yaml` in `internal/siteconfig`,
scaffolded projects receive env handling from nothing, and upcoming fullstack
apps would each re-implement precedence and struct binding.

`config` is a Runvil Libraries package providing one shared, framework-agnostic
loading model: a YAML file bound to a typed struct via `yaml` tags, overridden
by environment variables via `env` tags, with a fixed precedence. It never
imports `framework` or `net/http`, so any Go program in the ecosystem can use
it.

## 2. Problem Statement

Configuration in the Runvil ecosystem is duplicated and inconsistent:

- Every project reimplements "read file, overlay env, bind to struct".
- No common precedence rule means LDAP-style surprises: env flags, file
  values, and defaults combine differently per project.
- Struct-focused configuration (not `map[string]any`) keeps settings typed,
  validated, and documented, but requires a shared loader to be practical.

The result: config code is copy-pasted, precedence is ad hoc, and settings
drift out of sync with documentation. `config` removes that duplication.

## 3. Goals

- G1 — Bind configuration to typed structs using `yaml` and `env` tags.
- G2 — Overlay sources in a fixed precedence order.
- G3 — Treat a missing file as an empty config, never an error.
- G4 — Remain framework-agnostic: no `framework` imports, no HTTP concerns.
- G5 — Stay stdlib-first; add only `gopkg.in/yaml.v3` for YAML decoding.
- G6 — Produce deterministic, wrapped, debuggable errors.

## 4. Non-Goals

- NG1 — Config file watching or hot reload.
- NG2 — More serialization formats beyond YAML in this phase.
- NG3 — Remote secrets managers or encryption.
- NG4 — Config validation rules (see RVL-B9TW2); precedence is separate.
- NG5 — Defaults expressed *outside* the target struct (zero values are the base).

## 5. Requirements

### 5.1 Package & Loading

| ID          | Requirement                                                       | Priority |
| ----------- | ----------------------------------------------------------------- | -------- |
| CFG-LD-001  | Provide package `config` in `github.com/runvil/libs/config`.      | Must     |
| CFG-LD-002  | Provide `Load(path string, dst any) error` reading a YAML file and unmarshalling into `dst` (a pointer to a struct). | Must   |
| CFG-LD-003  | A missing file yields a zero `dst` and no error; a malformed file yields a wrapped error naming the path. | Must |
| CFG-LD-004  | File values bind by `yaml` tag name; a field without a tag binds by its lowercased field name (mirroring `yaml.v3`). | Must |
| CFG-LD-005  | Provide `Override(dst any, lookup func(string) (string, bool)) error` applying environment values on top of an already-loaded struct. | Must |

### 5.2 Environment Overlay

| ID          | Requirement                                                       | Priority |
| ----------- | ----------------------------------------------------------------- | -------- |
| CFG-EN-001  | A field configured with `env:"SERVER_ADDR"` is overridden from the environment using that exact key. | Must |
| CFG-EN-002  | A field without an `env` tag derives its key from `yaml` tag/name, uppercased with `-`/`.` replaced by `_` and the package prefix applied (e.g. `APP_TITLE`). | Should |
| CFG-EN-003  | The environment prefix (`config.WithPrefix`) overrides the default package prefix for implicit keys. | Should |
| CFG-EN-004  | Empty environment values are treated as unset (no override).      | Must |
| CFG-EN-005  | Nested struct fields recurse; path segments join with `_`.        | Must |

### 5.3 Precedence & Semantics

| ID          | Requirement                                                       | Priority |
| ----------- | ----------------------------------------------------------------- | -------- |
| CFG-PR-001  | Precedence is strictly: built-in zero values < file < environment. Later sources win. | Must |
| CFG-PR-002  | Slices and maps decode from YAML; slice elements are not environment-overlayed in this phase. | Must |
| CFG-PR-003  | An env value failing to decode into its field's type produces a wrapped error naming field and key. | Must |
| CFG-PR-004  | `Load` never panics on nil `dst`; it returns a usage-style wrapped error. | Must |

### 5.4 Org & Testing

| ID          | Requirement                                                       | Priority |
| ----------- | ----------------------------------------------------------------- | -------- |
| CFG-OR-001  | Provide `config.LoadOrDefault(path, dst) error` used by tooling that tolerates absent files. | Must |
| CFG-OR-002  | Every loading rule has an `*_test.go` verifying binding, precedence, and error paths. | Must |

## 6. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Dependencies.** Only Go stdlib + `gopkg.in/yaml.v3`.
- NFR3 — **Portability.** Linux, macOS, Windows.
- NFR4 — **Quality gates.** `gofmt`, `go vet ./...`, `go test ./...` pass.
- NFR5 — **Determinism.** Identical file + env yield identical structs.

## 7. Success Criteria

- S1 — A struct with `yaml` + `env` tags loads identically in a CLI tool, a
      server, and a test.
- S2 — Precedence test proves env wins over file, file wins over zero value.
- S3 — Removing the config file produces a zero config without error.
- S4 — `config` imports no `framework` package and no HTTP code.

## 8. Related Specifications

| SpecID    | Title                                           |
| --------- | ----------------------------------------------- |
| [RVL-4Y8UP](https://github.com/runvil/libs/blob/main/docs/specs/RVL-4Y8UP-runvil-libraries.md) | Runvil Libraries — Initial Specification |
| [RVL-B9TW2](./RVL-B9TW2-struct-validation.md) | Struct Validation |
| [RVN-K2SQ7](https://github.com/runvil/runvil/blob/main/docs/specs/RVN-K2SQ7-runvil-run-dev-deploy.md) | runvil run/dev/deploy (consumer) |

## 9. References

- [RVF-8G3WQ](https://github.com/runvil/framework/blob/main/docs/specs/RVF-8G3WQ-runvil-web-framework.md) — Runvil Web Framework (consumer).
- `gopkg.in/yaml.v3` — YAML library used for file decoding.
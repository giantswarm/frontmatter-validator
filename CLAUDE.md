## Project overview

An opinionated Hugo frontmatter validator for Giant Swarm's documentation. It scans Markdown files, extracts YAML frontmatter, and runs configurable validation checks. Outputs colored stdout, JSON, or GitHub Actions annotations.

## Build, test, lint

```bash
go run .                        # Run directly (preferred over building)
go test -race ./...             # Run all tests
go test -race ./pkg/validator/  # Run tests for a single package
golangci-lint run -E gosec -E goconst --timeout=15m ./...  # Lint
goimports -local github.com/giantswarm/frontmatter-validator -w .  # Fix imports
go mod tidy                     # After go.mod changes
make build                      # Build binary (uses architect for version)
```

Pre-commit hooks run `go-fmt`, `go-mod-tidy`, `golangci-lint`, and `go-imports` automatically.

## Architecture

Single-command CLI (cobra) with three packages:

- **`cmd/`** - CLI entry point. Handles flag parsing (`--output`, `--path`, `--config`), file discovery (walk dir or stdin), and orchestrates validation.
- **`pkg/validator/`** - Core validation logic.
  - `types.go` - All type definitions: `FrontMatter` struct (YAML tags), `Check`/`CheckResult`/`ValidationResult` types, check ID constants, `FlexibleDate` with custom YAML unmarshaling, runbook types.
  - `checks.go` - Check registry (`GetChecks()`) returning all checks with severity/description, and valid frontmatter keys whitelist (`GetValidKeys()`).
  - `validator.go` - `Validator` struct with `ValidateFile()` entry point. Uses `ConfigManager` interface to determine which checks are enabled per file path. Individual `validate*` methods for each field category.
- **`pkg/config/`** - YAML config loading with path-based check filtering. `Manager` resolves enabled checks per file using `default_rules` + `directory_overrides` with glob matching.
- **`pkg/output/`** - `Formatter` for stdout (colored), JSON, and GitHub Actions annotation output. Uses `afero` for testable filesystem operations.

## Key conventions

- Maintain `CHANGELOG.md` using [Keep a Changelog](https://keepachangelog.com/) format for all changes.
- Keep `frontmatter-validator.schema.json` in sync with any changes to checks or configuration structure.
- Prefer table-style unit tests.
- Sort imports with `goimports -local github.com/giantswarm/frontmatter-validator`.
- Use sentence case for documentation headlines.
- Check IDs are `UPPER_SNAKE_CASE` string constants defined in `pkg/validator/types.go`.
- Adding a new check requires: constant in `types.go`, entry in `GetChecks()` in `checks.go`, validation logic in `validator.go`, and adding the ID to the default enabled list in both `validator.go` (`defaultConfigManager`) and `config/manager.go` (`getDefaultConfig`).

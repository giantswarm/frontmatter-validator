# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).



## [Unreleased]

### Added

- Added `ignore_paths` configuration option to completely skip files from validation. Supports glob patterns (e.g., `vendor/**`) and exact paths (e.g., `README.md`).

## [0.4.0] - 2026-03-10

### Added

- Support usage as a [pre-commit](https://pre-commit.com/) hook. The validator now accepts file paths as positional arguments and exits with a non-zero code when validation issues are found.

### Changed

- Migrate from gopkg.in/yaml.v3 to go.yaml.in/yaml/v4

## [0.3.2] - 2026-03-03

### Fixed

- Fixed `NO_TITLE`, `SHORT_TITLE`, `LONG_TITLE`, `UNKNOWN_ATTRIBUTE`, `NO_WEIGHT`, and `NO_LINK_TITLE` checks ignoring `disabled_checks` configuration. These checks were not consulting the config manager, so `directory_overrides` had no effect on them.

## [0.3.1] - 2025-10-01

### Fixed

- Fixed front matter parsing to handle quoted date strings in `last_review_date` field. The validator now supports various date formats including quoted strings (`"2025-01-10"`), unquoted dates (`2025-01-10`), and full timestamps (`"2025-01-10T00:00:00Z"`). This resolves false `NO_FRONT_MATTER` errors when files contain valid front matter with quoted dates.

## [0.3.0] - 2025-09-30

### Added

- Added checks related to runbooks.

### Removed

- Removed `--validation` flag. This functionality can now be covered by multiple configuration files.

## [0.2.0] - 2025-09-30

### Added

- Add configuration system with YAML-based rule management for directory-specific validation rules
- Add `--config` flag to specify configuration file location (default: `./frontmatter-validator.yaml`)
- Add example configuration files for full validation and last-reviewed-only modes
- Add JSON Schema (`frontmatter-validator.schema.json`) for IDE validation and autocompletion support

### Removed

- Removed hardcoded ignore paths in favour of the flexible configuration system
- Removed `--exclude` flag - path-specific exclusions should now be handled via configuration files

### Fixed

- Fixed pluralization

## [0.1.1] - 2025-09-29

### Fixed

- Fix STDIN input validation to accept any `.md` file paths regardless of `--path` flag
- Remove incorrect targetPath prefix check that was filtering valid file paths from STDIN

### Added

- Add comprehensive test suite for STDIN input handling in `cmd/root_test.go`
- Add virtual filesystem tests for annotations JSON writing using Afero library
- Add `DumpAnnotationsToFS` method to support testable filesystem operations

## [0.1.0] - 2025-09-29

### Changed

- Add validation for `--validation` flag value

## [0.0.1] - 2025-09-29

### Added

- First implementation

[Unreleased]: https://github.com/giantswarm/frontmatter-validator/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/giantswarm/frontmatter-validator/compare/v0.3.2...v0.4.0
[0.3.2]: https://github.com/giantswarm/frontmatter-validator/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/giantswarm/frontmatter-validator/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/frontmatter-validator/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/frontmatter-validator/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/giantswarm/frontmatter-validator/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/giantswarm/frontmatter-validator/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/giantswarm/frontmatter-validator/releases/tag/v0.0.1

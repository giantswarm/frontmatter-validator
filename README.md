# Frontmatter Validator

An opinionated validator for HUGO frontmatter, designed for Giant Swarm requirements.

Frontmatter is metadata enclosed in Markdown files, like page title, description, and more.

## Features

- Written in Go
- Parses YAML frontmatter using `gopkg.in/yaml.v3`
- Configurable set of validation rules
- Creates GitHub Actions run annotations for problems found
- Multiple output formats (stdout with colors, JSON)
- Supports validation modes (all checks or last-review-date only)
- Command-line interface with flexible input options

## Logic

- Scan a target path recursively for directories and files.
- For each file, check if the file matches a configured pattern. By default, the pattern `*.md` is used to match Markdown files.
- For each matched file:
  - Check if there is a frontmatter block in the file, wrapped in `\n---`. If not, issue a `NO_FRONT_MATTER` error and skip the file.
  - Read and parse the frontmatter enclosed in the `---` wrapper in the head of the Markdown file.
  - Issue an error and skip the file if the frontmatter does not parse as valid YAML.
  - Apply all validators that are configured for the given path. By default, all validators should be applied.
  - Yield warnings and errors as annotations and log lines.

## Validators

Validator names are given as the short form of the "complaint" they issue, so they are all meant in a negative way. Sorry 'bout that! `;-)`

These validator IDs can be used in configuration files to enable or disable specific checks for different directories.

### General

- `NO_FRONTMATTER`: there is no frontmatter at all.
- `UNKNOWN_ATTRIBUTE`: the frontmatter contains an attribute that is not in the list of valid keys.
- `NO_TRAILING_NEWLINE`: the file does not end with a newline character, which is required for proper parsing.

### Title

- `NO_TITLE`: the `title` field is missing or empty.
- `LONG_TITLE`: the `title` is longer than 100 characters.
- `SHORT_TITLE`: the `title` is shorter than 5 characters.

### Description

- `INVALID_DESCRIPTION`: the `description` is not a string or contains line breaks.
- `LONG_DESCRIPTION`: the `description` is longer than 300 characters.
- `NO_FULL_STOP_DESCRIPTION`: the `description` does not end with a full stop.
- `NO_DESCRIPTION`: the `description` field is missing.
- `SHORT_DESCRIPTION`: the `description` is shorter than 50 characters.

### Owner

- `INVALID_OWNER`: the `owner` is not an array of valid GitHub team URLs starting with `https://github.com/orgs/giantswarm/teams/`.
- `NO_OWNER`: the `owner` field is missing or empty.

### Last review date

- `INVALID_LAST_REVIEW_DATE`: the `last_review_date` is not a valid date in the form `YYYY-MM-DD` or is set to a future date.
- `NO_LAST_REVIEW_DATE`: the `last_review_date` field is missing.
- `REVIEW_TOO_LONG_AGO`: the `last_review_date` is older than the expiration period (default 365 days, configurable via `expiration_in_days`).

### Link title

- `LONG_LINK_TITLE`: the `linkTitle` (or `title` if `linkTitle` is not provided) is longer than 40 characters.
- `NO_LINK_TITLE`: the `linkTitle` field is missing when the page has a menu configuration.

### Weight

- `NO_WEIGHT`: the `weight` field is missing when the page has a menu configuration.

### User questions

- `NO_USER_QUESTIONS`: the `user_questions` field is missing (except for `_index.md` files).
- `LONG_USER_QUESTION`: any user question is longer than 100 characters.
- `NO_QUESTION_MARK`: a user question does not end with a question mark.

## Installation and usage

### Building from Source

```bash
# Clone the repository
git clone https://github.com/giantswarm/frontmatter-validator
cd frontmatter-validator

# Build the binary
go build -o frontmatter-validator .
```

### Command line usage

```bash
# Validate all markdown files in current directory (default)
./frontmatter-validator

# Validate files in a specific directory
./frontmatter-validator --path=/path/to/docs

# Only validate last review dates
./frontmatter-validator --validation=last-reviewed

# Output results as JSON (useful for CI/CD integration)
./frontmatter-validator --output=json

# Validate specific files via stdin
echo "src/content/docs/example.md" | ./frontmatter-validator

# Show help
./frontmatter-validator --help
```

### Available flags

- `--validation`: Validation mode (`all` or `last-reviewed`, default: `all`)
- `--output`: Output format (`stdout` or `json`, default: `stdout`)
- `--path`: Target path to scan for Markdown files (default: `.`)
- `--config`: Path to configuration file (default: `./frontmatter-validator.yaml`)

## Configuration

The frontmatter validator supports flexible configuration through YAML files. This allows you to define which validation checks are enabled for different directories, making it easy to have different validation rules for different types of content.

### Configuration file location

By default, the validator looks for `./frontmatter-validator.yaml` in the current directory. You can specify a different configuration file using the `--config` flag:

```bash
# Use default configuration file
./frontmatter-validator

# Use a specific configuration file
./frontmatter-validator --config=./my-custom-config.yaml

# Use configuration for last-reviewed validation mode
./frontmatter-validator --config=./frontmatter-validator-last-reviewed.yaml
```

### Configuration file format

The configuration file uses YAML format with two main sections:

```yaml
# Default rules applied to all files
default_rules:
  enabled_checks:
    - NO_TITLE
    - NO_DESCRIPTION
    - NO_OWNER
    # ... more checks

# Directory-specific overrides
directory_overrides:
  - path: "src/content/vintage/**"
    disabled_checks:
      - NO_DESCRIPTION
      - NO_OWNER
  
  - path: "src/content/changes/**"
    disabled_checks:
      - NO_USER_QUESTIONS
```

### Configuration sections

#### `default_rules`
Defines the baseline validation rules that apply to all files unless overridden.

- `enabled_checks`: List of validation check IDs that should be enabled by default

#### `directory_overrides`
Allows you to override the default rules for specific directory patterns.

- `path`: Glob pattern matching file paths (e.g., `src/content/vintage/**`)
- `enabled_checks`: Additional checks to enable for this path (optional)
- `disabled_checks`: Checks to disable for this path (optional)

### Path Patterns

Directory overrides support glob patterns:

- `src/content/vintage/**` - Matches all files under the vintage directory and subdirectories
- `src/content/changes/*` - Matches files directly under changes directory (not subdirectories)
- `**/_index.md` - Matches all `_index.md` files anywhere in the tree
- `src/content/docs/example.md` - Matches a specific file

### Available Check IDs

Use the exact check IDs from the [Validators](#validators) section in your configuration files. All validator IDs can be referenced in the `enabled_checks` and `disabled_checks` lists.

### Example configurations

#### Full validation (default)

```yaml
default_rules:
  enabled_checks:
    - NO_TITLE
    - NO_DESCRIPTION
    - NO_OWNER
    - NO_LAST_REVIEW_DATE
    # ... all other checks

directory_overrides:
  - path: "src/content/vintage/**"
    disabled_checks:
      - NO_LAST_REVIEW_DATE
      - REVIEW_TOO_LONG_AGO
```

#### Last review date only

```yaml
default_rules:
  enabled_checks:
    - NO_FRONT_MATTER
    - NO_TRAILING_NEWLINE
    - NO_LAST_REVIEW_DATE
    - REVIEW_TOO_LONG_AGO
    - INVALID_LAST_REVIEW_DATE

directory_overrides:
  - path: "src/content/vintage/**"
    disabled_checks:
      - NO_LAST_REVIEW_DATE
      - REVIEW_TOO_LONG_AGO
```

### Migration from --validation Flag

The `--validation` flag is now deprecated in favor of configuration files:

- Instead of `--validation=all`, use the default configuration or `--config=./frontmatter-validator.yaml`
- Instead of `--validation=last-reviewed`, use `--config=./frontmatter-validator-last-reviewed.yaml`


### GitHub Actions integration

When running in GitHub Actions (detected via `GITHUB_ACTIONS` environment variable), the validator automatically creates an `annotations.json` file that can be used with the [annotations-action](https://github.com/yuzutech/annotations-action) to display validation results as PR annotations.

### Output formats

#### Standard output (default)

Provides colored output with severity levels:

- 🔴 **FAIL**: Critical issues that must be fixed
- 🟡 **WARN**: Less severe issues that should be addressed

#### JSON Output

Structured output suitable for integration with issue tracking systems and CI/CD pipelines.

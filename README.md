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

## Checks

See [documentation / Checks](docs/checks.md) for information on the checks that can be performed.

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

# Output results as JSON (useful for CI/CD integration)
./frontmatter-validator --output=json

# Validate specific files via stdin
echo "src/content/docs/example.md" | ./frontmatter-validator

# Show help
./frontmatter-validator --help
```

### Available flags

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

The configuration file uses YAML format with two main sections. A JSON Schema is provided at `frontmatter-validator.schema.json` for IDE validation and autocompletion support.

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

Use the exact check IDs from the [Checks](#checks) section in your configuration files. All validator IDs can be referenced in the `enabled_checks` and `disabled_checks` lists.

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

### GitHub Actions integration

When running in GitHub Actions (detected via `GITHUB_ACTIONS` environment variable), the validator automatically creates an `annotations.json` file that can be used with the [annotations-action](https://github.com/yuzutech/annotations-action) to display validation results as PR annotations.

### Output formats

#### Standard output (default)

Provides colored output with severity levels:

- 🔴 **FAIL**: Critical issues that must be fixed
- 🟡 **WARN**: Less severe issues that should be addressed

#### JSON Output

Structured output suitable for integration with issue tracking systems and CI/CD pipelines.

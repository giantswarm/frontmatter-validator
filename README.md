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

### General

- NO_FRONTMATTER: there is no frontmatter at all.
- UNKNOWN_ATTRIBUTE: the frontmatter contains an attribute that is not in the list of valid keys.
- NO_TRAILING_NEWLINE: the file does not end with a newline character, which is required for proper parsing.

### Title

- NO_TITLE: the `title` field is missing or empty.
- LONG_TITLE: the `title` is longer than 100 characters.
- SHORT_TITLE: the `title` is shorter than 5 characters.

### Description

- INVALID_DESCRIPTION: the `description` is not a string or contains line breaks.
- LONG_DESCRIPTION: the `description` is longer than 300 characters.
- NO_FULL_STOP_DESCRIPTION: the `description` does not end with a full stop.
- NO_DESCRIPTION: the `description` field is missing.
- SHORT_DESCRIPTION: the `description` is shorter than 50 characters.

### Owner

- INVALID_OWNER: the `owner` is not an array of valid GitHub team URLs starting with `https://github.com/orgs/giantswarm/teams/`.
- NO_OWNER: the `owner` field is missing or empty.

### Last review date

- INVALID_LAST_REVIEW_DATE: the `last_review_date` is not a valid date in the form `YYYY-MM-DD` or is set to a future date.
- NO_LAST_REVIEW_DATE: the `last_review_date` field is missing.
- REVIEW_TOO_LONG_AGO: the `last_review_date` is older than the expiration period (default 365 days, configurable via `expiration_in_days`).

### Link title

- LONG_LINK_TITLE: the `linkTitle` (or `title` if `linkTitle` is not provided) is longer than 40 characters.
- NO_LINK_TITLE: the `linkTitle` field is missing when the page has a menu configuration.

### Weight

- NO_WEIGHT: the `weight` field is missing when the page has a menu configuration.

### User questions

- NO_USER_QUESTIONS: the `user_questions` field is missing (except for `_index.md` files).
- LONG_USER_QUESTION: any user question is longer than 100 characters.
- NO_QUESTION_MARK: a user question does not end with a question mark.

## Installation and Usage

### Building from Source

```bash
# Clone the repository
git clone https://github.com/giantswarm/frontmatter-validator
cd frontmatter-validator

# Build the binary
go build -o frontmatter-validator .
```

### Command Line Usage

```bash
# Validate all markdown files in src/content directory (default)
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

### Available Flags

- `--validation`: Validation mode (`all` or `last-reviewed`, default: `all`)
- `--output`: Output format (`stdout` or `json`, default: `stdout`)
- `--path`: Target path to scan for Markdown files (default: `src/content`)
- `--exclude`: Exclude paths from checks (can be used multiple times)

### Excluding Paths from Validation

The `--exclude` flag allows you to skip validation for specific paths and checks. This is useful for legacy content, generated files, or special cases that don't need to follow all validation rules.

#### Exclude Formats

```bash
# Exclude all checks for a specific path
--exclude="path/to/exclude"

# Exclude specific check(s) for a path
--exclude="path/to/exclude:CHECK_NAME"

# Exclude multiple checks for a path
--exclude="path/to/exclude:CHECK1,CHECK2"

# Use multiple exclude flags
--exclude="legacy/" --exclude="docs/old:NO_DESCRIPTION"
```

#### Examples

```bash
# Skip all validation for legacy documentation
./frontmatter-validator --exclude="docs/legacy"

# Skip description validation for generated files
./frontmatter-validator --exclude="docs/generated:NO_DESCRIPTION,LONG_DESCRIPTION"

# Skip multiple checks for different paths
./frontmatter-validator \
  --exclude="docs/legacy" \
  --exclude="docs/api:NO_USER_QUESTIONS" \
  --exclude="README.md:NO_FRONT_MATTER"

# Complex exclusion example
./frontmatter-validator \
  --path="src/content" \
  --exclude="src/content/vintage" \
  --exclude="src/content/changes:NO_DESCRIPTION,NO_USER_QUESTIONS" \
  --exclude="src/content/reference/platform-api/crd:NO_OWNER"
```

#### Path Matching

- **Exact match**: `README.md` matches only `README.md`
- **Directory match**: `docs/legacy` matches `docs/legacy/file.md` and `docs/legacy/sub/file.md`
- **Flexible paths**: Both `./README.md` and `README.md` patterns work
- **Case sensitive**: Path matching is case sensitive

#### Available Check Names

Use the exact check names from the validator output (e.g., `NO_DESCRIPTION`, `LONG_TITLE`, `INVALID_OWNER`). See the [Validators](#validators) section for the complete list.

### GitHub Actions Integration

When running in GitHub Actions (detected via `GITHUB_ACTIONS` environment variable), the validator automatically creates an `annotations.json` file that can be used with the [annotations-action](https://github.com/yuzutech/annotations-action) to display validation results as PR annotations.

### Output Formats

#### Standard Output (default)
Provides colored output with severity levels:
- 🔴 **FAIL**: Critical issues that must be fixed
- 🟡 **WARN**: Less severe issues that should be addressed

#### JSON Output

Structured output suitable for integration with issue tracking systems and CI/CD pipelines.

# Test Data

This directory contains test files used by the validator tests.

## Structure

### runbooks/

Contains test files for runbook validation:

#### Valid Cases
- `valid-complete.md` - Complete runbook with all sections populated
- `valid-minimal.md` - Minimal runbook with empty runbook config
- `valid-empty-arrays.md` - Runbook with empty arrays for all optional sections
- `valid-only-variables.md` - Runbook with only variables defined

#### Invalid Cases
- `layout-not-set.md` - Runbook config without `layout: runbook`
- `appears-in-menu.md` - Runbook without `toc_hide: true`
- `variable-without-name.md` - Variable missing required `name` field
- `invalid-variable-name.md` - Variable with invalid name format
- `duplicate-variable-names.md` - Multiple variables with same name
- `dashboard-without-name.md` - Dashboard missing required `name` field
- `dashboard-without-link.md` - Dashboard missing required `link` field
- `dashboard-undefined-variable.md` - Dashboard link uses undefined variable
- `dashboard-invalid-url.md` - Dashboard with invalid URL format
- `known-issue-without-url.md` - Known issue missing required `url` field
- `known-issue-invalid-url.md` - Known issue with invalid URL format

## Adding New Test Cases

To add a new test case:

1. Create a new `.md` file in the appropriate subdirectory
2. Add the test case to the table-driven test in `validator_test.go`
3. Specify either `expectValid: true` for valid cases or `expectChecks: []string{...}` for invalid cases

## Benefits of This Approach

- **Maintainability**: Test inputs are separate from test logic
- **Readability**: Easy to see what each test case is testing
- **Extensibility**: Adding new test cases is simple
- **Reusability**: Test files can be used for manual testing or other tools
- **Version Control**: Changes to test cases are clearly visible in diffs

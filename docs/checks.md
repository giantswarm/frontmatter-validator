# Checks

Here you find information about all the checks frontmatter-validator can perform.

Regarding naming: check names are given as the short form of the "complaint" they yield, in uppercase letters, using underscore as separator. These check names are used in configuration files to enable or disable specific checks for different directories.

### General

- `NO_FRONTMATTER`: checks whether there is frontmatter in the file. If this error occurs, it means that there is no frontmatter at all.
- `UNKNOWN_ATTRIBUTE`: checks whether there are any unknown attributes. If this error occurs, the frontmatter contains an attribute that is not in the list of valid keys.
- `NO_TRAILING_NEWLINE`: Checks whether the file ends in a newlinw (which is required for proper parsing). If this error occurs, the file does not end with a newline character.

### Title

- `NO_TITLE`: checks whether the `title` field is present and non-empty.
- `LONG_TITLE`: checks whether the `title` value is longer than 100 characters.
- `SHORT_TITLE`: checks whether the `title` value is shorter than 5 characters.

### Description

- `NO_DESCRIPTION`: checks whether the `description` field is present and non-empty.
- `INVALID_DESCRIPTION`: checks whether the `description` is a string and doesn't contain line breaks.
- `LONG_DESCRIPTION`: checks if the `description` is longer than 300 characters.
- `SHORT_DESCRIPTION`: checks if the `description` is shorter than 50 characters.
- `NO_FULL_STOP_DESCRIPTION`: checks if the `description` ends with a full stop.

### Owner

- `NO_OWNER`: checks if the `owner` field is present and not empty.
- `INVALID_OWNER`: checks if the `owner` is an array of valid GitHub team URLs starting with `https://github.com/orgs/giantswarm/teams/`.

### Last review date

- `NO_LAST_REVIEW_DATE`: checks if the `last_review_date` field is present.
- `INVALID_LAST_REVIEW_DATE`: checks if the `last_review_date` is a valid date in the past in the form `YYYY-MM-DD`.
- `REVIEW_TOO_LONG_AGO`: checks if the `last_review_date` is older than the expiration period (default 365 days, configurable via the `expiration_in_days` frontmatter field).

### Link title

- `NO_LINK_TITLE`: checks if the page has a menu configuration AND the `linkTitle` field is present.
- `LONG_LINK_TITLE`: checks if the title used in the menu (either `linkTitle` if present, otherwise `title`) is shorter than 40 characters.

### Weight

- `NO_WEIGHT`: checks if the `weight` field is present when the page has a menu configuration.

### Runbooks

Devops runbooks require their own group of frontmatter fields under the `runbook` key. Example:

```yaml
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
      default: golem
  dashboards:
    - name: Cilium performance
      link: https://grafana-$INSTALLATION.teleport.giantswarm.io/d/d57506f1-ee2d-4f3e-8687-c8e1a610b8c6/cilium-performance
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/28493
      description: Version 1.13 performance is not great in general, e.g. when it comes to create a large number of pods when there are many CIDR-based policies in the cluster (see "Identities" graphs for more information).
    - url: https://github.com/giantswarm/giantswarm/issues/29998
      description: Performance of Hubble UI in version 1.13 and 1.14 isn't good, and we are not convinced 1.15 is that much better.
layout: runbook
toc_hide: true
```

Runbook checks:

- `RUNBOOK_LAYOUT_NOT_SET`: checks if the `layout: runbook` field is set.
- `INVALID_RUNBOOK_VARIABLES`: checks if `runbook.variables` is an array and isn't empty.
- `RUNBOOK_VARIABLE_WITHOUT_NAME`: checks whether each runbook variable has a name specified.
- `INVALID_RUNBOOK_VARIABLE_NAME`: checks whether a variable name is using only uppercase letters and the underscore. Also the name must be unique within this runbook's variables.
- `INVALID_RUNBOOK_VARIABLE`: checks whether each variable is a valid object with the field `name` and the optional fields `description` and `default`.
- `INVALID_RUNBOOK_DASHBOARDS`: checks if `runbook.dashboards` is an array and isn't empty.
- `INVALID_RUNBOOK_DASHBOARD`: checks whether each runbook dashboard has `name` and `link` specified and non-empty.
- `INVALID_RUNBOOK_DASHBOARD_LINK`: checks whether the dashboard `link` is a valid URL. This includes checking that the variables used like `$INSTALLATION` are defined in the runbook variables.
- `INVALID_RUNBOOK_KNOWN_ISSUES`: checks if `runbook.known_issues` is an array and isn't empty.
- `INVALID_RUNBOOK_KNOWN_ISSUE`: checks whether each known issue has `url` defined. The entry may also have the optional field `description`.
- `INVALID_RUNBOOK_KNOWN_ISSUE_URL`: checks whether a known issue URL is a valid URL.
- `RUNBOOK_APPEARS_IN_MENU`: checks whether the `toc_hide: true` field is set.


### User questions

- `NO_USER_QUESTIONS`: checks if the `user_questions` field is missing (except for `_index.md` files).
- `LONG_USER_QUESTION`: checks if any user question is longer than 100 characters.
- `NO_QUESTION_MARK`: checks if a user question does not end with a question mark.

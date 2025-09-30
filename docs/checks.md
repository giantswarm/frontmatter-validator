# Checks

Here you find information about all the checks frontmatter-validator can perform.

Regarding naming: check names are given as the short form of the "complaint" they yield, in uppercase letters, using underscore as separator. These check names are used in configuration files to enable or disable specific checks for different directories.

### General

- `NO_FRONTMATTER`: checks whether there is frontmatter in the file. If this error occurs, it means that there is no frontmatter at all.
- `UNKNOWN_ATTRIBUTE`: checks whether there are any unknown attributes. If this error occurs, the frontmatter contains an attribute that is not in the list of valid keys.
- `NO_TRAILING_NEWLINE`: Chceks whether the file ends in a newlinw (which is required for proper parsing). If this error occurs, the file does not end with a newline character.

### Title

- `NO_TITLE`: checks whether the `title` field is present and non-empty.
- `LONG_TITLE`: checks whether the `title` value is longer than 100 characters.
- `SHORT_TITLE`: checks whether the `title` value is shorter than 5 characters.

### Description

- `NO_DESCRIPTION`: cehcks whether the `description` field is present and non-empty.
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

### User questions

- `NO_USER_QUESTIONS`: checks if the `user_questions` field is missing (except for `_index.md` files).
- `LONG_USER_QUESTION`: checks if any user question is longer than 100 characters.
- `NO_QUESTION_MARK`: chceks if a user question does not end with a question mark.

---
title: Test Runbook With Multiple Errors
description: This is a valid runbook description that is long enough and ends with a full stop.
layout: runbook
# Missing toc_hide: true
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
runbook:
  variables:
    - name: invalid-name  # Invalid format
      description: Invalid name format
    - description: Missing name  # Missing name
  dashboards:
    - name: Test Dashboard
      link: not-a-valid-url  # Invalid URL
  known_issues:
    - description: Missing URL  # Missing URL
---

# Test Content

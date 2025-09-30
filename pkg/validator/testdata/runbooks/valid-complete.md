---
title: Valid Complete Runbook
description: This is a valid runbook description that is long enough and ends with a full stop.
layout: runbook
toc_hide: true
owner:
  - https://github.com/orgs/giantswarm/teams/team-honeybadger
last_review_date: 2024-09-01
runbook:
  variables:
    - name: INSTALLATION
      description: Installation name
      default: golem
    - name: CLUSTER_ID
      description: Cluster identifier
  dashboards:
    - name: Cilium performance
      link: https://grafana-$INSTALLATION.teleport.giantswarm.io/d/cilium-performance
  known_issues:
    - url: https://github.com/giantswarm/giantswarm/issues/28493
      description: Version 1.13 performance is not great in general
    - url: https://github.com/giantswarm/giantswarm/issues/29998
---

# Test Runbook Content

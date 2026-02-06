---
title: Use Cases
slug: user-guide/use-cases
section: User Guide
order: 50
tags:
  - user-guide
  - workflow
status: published
---
# Use Cases

Below are a few practical workflows you can adopt with `mochi-sticky`.

## Feature Branch Boards + Docs + ADRs

A lightweight workflow for shipping a feature while keeping tasks, docs, and decisions reviewable in Git.

### 1) Capture Feature Tasks On The Default Board

```bash
mochi-sticky task add "Create product CRUD" --tags feature,product --priority 1
mochi-sticky task add "Add product table migration" --tags feature,product --priority 1
```

### 2) Create A Branch And A Dedicated Feature Board

```bash
git checkout -b feat/product-crud
mochi-sticky board add "Feature: product CRUD"
```

### 3) Switch To The Feature Board And Split Into Implementation Tasks

```bash
mochi-sticky board list
mochi-sticky board use <feature-board-id>

mochi-sticky task add "API: create product" --tags api,product --priority 1
mochi-sticky task add "API: list products" --tags api,product --priority 2
mochi-sticky task add "UI: product form" --tags ui,product --priority 2
```

### 4) Document The Feature In The Wiki And Record ADRs

```bash
mochi-sticky wiki create "Product CRUD" --slug features/product-crud --tags product,feature --status draft
mochi-sticky wiki edit features/product-crud

mochi-sticky adr create "Product identifiers" --status proposed --tags product --links features/product-crud
```

### 5) Work The Tasks

```bash
mochi-sticky task list --sort priority
mochi-sticky task move <task-id> doing
mochi-sticky task move <task-id> done
```

### 6) Merge The PR And Clean Up The Feature Board

If you want to keep a record, archive first; otherwise delete directly.

```bash
mochi-sticky board archive <feature-board-id> --force
mochi-sticky board delete <feature-board-id> --force
```

This keeps the default board focused on high-level planning while feature boards stay scoped to a single branch/PR.

## Resolve A Production Issue (Bugfix Workflow)

A pragmatic workflow for taking an issue from report to fix, with traceability in tasks, docs, and ADRs when needed.

### 1) Capture The Issue As A Task

```bash
mochi-sticky task add "Fix 500 when creating product with empty name" --tags bug,product --priority 1
```

### 2) Create A Branch And A Focused Board (Optional)

If the issue is complex or will span multiple PRs, spin up a dedicated board.

```bash
git checkout -b fix/product-create-500
mochi-sticky board add "Bug: product create 500"
mochi-sticky board use <bug-board-id>
```

### 3) Break Down Investigation + Fix + Verification

```bash
mochi-sticky task add "Reproduce locally + add failing test" --tags bug,test --priority 1
mochi-sticky task add "Implement validation + error mapping" --tags bug,api --priority 1
mochi-sticky task add "Add regression test coverage" --tags bug,test --priority 1
mochi-sticky task add "Update troubleshooting docs" --tags docs,runbook --priority 2
```

### 4) Document The Fix (And Decisions)

```bash
mochi-sticky wiki create "Troubleshooting: Product Create 500" --slug troubleshooting/product-create-500 --tags troubleshooting,product --status draft
mochi-sticky wiki edit troubleshooting/product-create-500

# If you need to capture a decision (e.g., status codes, validation rules)
mochi-sticky adr create "Product create validation rules" --status proposed --tags product --links troubleshooting/product-create-500
```

### 5) Close Out And Clean Up

```bash
mochi-sticky task list --sort priority
mochi-sticky task move <task-id> done

mochi-sticky board archive <bug-board-id> --force
mochi-sticky board delete <bug-board-id> --force
```

## Plan A Milestone (Release Planning)

Use a board as a milestone plan, track scope, and keep release docs close to the work.

### 1) Create A Milestone Board

```bash
mochi-sticky board add "Milestone: v0.3.0"
mochi-sticky board use <milestone-board-id>
```

### 2) Add Epics / Themes As Tasks

```bash
mochi-sticky task add "CLI polish + help text" --tags milestone,v0.3.0,cli --priority 2
mochi-sticky task add "Wiki export improvements" --tags milestone,v0.3.0,wiki --priority 1
mochi-sticky task add "Stability + regression fixes" --tags milestone,v0.3.0,stability --priority 1
```

### 3) Track Decisions And Notes In The Wiki

```bash
mochi-sticky wiki create "Release Plan: v0.3.0" --slug releases/v0.3.0/plan --tags release,planning --status draft
mochi-sticky wiki edit releases/v0.3.0/plan
```

### 4) Use ADRs For Breaking Changes Or Major Decisions

```bash
mochi-sticky adr create "v0.3.0 breaking changes policy" --status proposed --tags release --links releases/v0.3.0/plan
```

### 5) Execution And Wrap-Up

```bash
mochi-sticky task list --sort priority
mochi-sticky task ready

# After shipping, you can archive the milestone board to keep history.
mochi-sticky board archive <milestone-board-id> --force
```

---
title: PR Description Quality Standards
description: Ensures PR descriptions meet quality criteria for the cadence-samples (Go) repo using guidance from the PR template and .github/pull_request_guidance.md
when: PR description is created or updated
actions: Read PR template and guidance, then report requirement status
---

# PR Description Quality Standards

When evaluating a pull request description:

1. **Read the PR template guidance** at `.github/pull_request_guidance.md` to understand the expected guidance for each section
2. Apply that guidance to evaluate the current PR description
3. Provide recommendations for how to improve the description.

## Core Principle: Why Not How

From https://cbea.ms/git-commit/#why-not-how:
- **"A diff shows WHAT changed, but only the description can explain WHY"**
- Focus on: the problem being solved, the reasoning behind the solution, context
- The code itself documents HOW - the PR description documents WHY

## Evaluation Criteria

### Required Sections (must exist with substantive content per PR template guidance)

1. **Which sample(s) or area?** 
   - One line listing area(s) touched. This repo has two sample trees: **cmd/samples** (legacy; make + ./bin/…) and **new_samples** (per-folder; go run .). Identify which tree and area(s) are touched.
   - **cmd/samples:** e.g. cmd/samples/recipes/helloworld, cmd/samples/recipes/branch, cmd/samples/recipes/query, cmd/samples/batch, cmd/samples/cron, cmd/samples/expense, cmd/samples/fileprocessing, cmd/samples/dsl, cmd/samples/pso, cmd/samples/recovery, cmd/samples/recipes/cancelactivity, etc.
   - **new_samples:** hello_world, activities, query, signal, operations, client_tls, template.
   - **Other:** README, build, config, Makefile, common.
   - Helps reviewers route; skip flagging if area is obvious from paths

2. **What changed?**
   - 1-2 line summary of WHAT changed technically
   - Focus on key modification, not implementation details
   - Link to GitHub issue encouraged for non-trivial changes; optional for trivial doc/sample tweaks
   - Template has good/bad examples

3. **Why?**
   - Context and motivation (why not how)
   - Enough rationale for reviewers to understand the goal (e.g. improving clarity, fixing compatibility, aligning with docs)
   - Must explain WHY this approach was chosen

4. **How did you test it?**
   - Concrete, copyable commands with exact invocations
   - GOOD: `make`, `go test ./...`, `go test ./cmd/samples/recipes/helloworld/`, `cd new_samples/hello_world && go run .`, and Cadence CLI commands as in sample READMEs
   - BAD: "Tested locally" or "See tests"
   - Expect Go/make and/or sample execution commands; no canary or integration server setup required

5. **Potential risks**
   - Often N/A for sample-only or doc-only changes
   - Call out when relevant: dependency upgrades, behavior changes for someone copying the sample, build/config changes
   - Don't require lengthy text when N/A is appropriate

6. **Release notes**
   - Optional for this repo. Use when change is user-facing (e.g. new sample, notable README change)
   - Can be N/A for internal refactors or tiny fixes
   - Don't require lengthy text when N/A is appropriate

7. **Documentation Changes**
   - Often relevant when adding or changing samples (main README, cmd/samples READMEs, new_samples READMEs including generator-generated, links to cadence or cadence-docs)
   - Only mark N/A if certain no docs are affected

### Quality Checks

- **Skip obvious things** - Don't flag items clear from folder structure (e.g. area from paths)
- **Skip trivial refactors** - Minor formatting/style changes don't need deep rationale
- **Don't check automated items** - CI, linting are automated

## FORBIDDEN - Never Include

- "Issues Found", "Testing Evidence Quality", "Documentation Reasoning", "Summary" sections
- "Note:" paragraphs or explanatory text outside recommendations
- Grouping recommendations by type

## Section Names (Use EXACT Brackets)

- **[Which sample(s) or area?]**
- **[What changed?]**
- **[Why?]**
- **[How did you test it?]**
- **[Potential risks]**
- **[Release notes]**
- **[Documentation Changes]**

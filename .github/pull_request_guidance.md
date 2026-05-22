<!-- List the area(s) touched so reviewers know where to look.
This repo has two sample trees: legacy samples under cmd/samples (built with make, run via ./bin/<name>) and newer samples under new_samples (per-folder, run with go run .). Naming the area helps reviewers.
Examples: cmd/samples/recipes/helloworld, cmd/samples/batch, cmd/samples/expense, new_samples/hello_world, new_samples/query, new_samples/operations, README, build, config, Makefile -->
**Which sample(s) or area?**


<!-- 1-2 line summary of WHAT changed technically.
- Link to a GitHub issue when applicable (encouraged for larger changes; optional for trivial doc/sample tweaks)
- Good: "Added CancelWorkflow sample in new_samples/operations" or "Updated README run instructions for Go 1.21"
- Bad: "updated code" or "fixed stuff" -->
**What changed?**


<!-- Provide context and motivation (see https://cbea.ms/git-commit/#why-not-how). Focus on WHY, not how.
- Lighter than core Cadence repo: e.g. "improving clarity for new users," "fixing sample broken on Go 1.21," "aligning with cadence-docs"
- Still give enough rationale for reviewers to understand the goal
- Good: "HelloWorld didn't show retry behavior; this sample demonstrates RetryOptions so users can copy-paste a working example."
- Bad: "Improves samples" -->
**Why?**


<!-- Include concrete, copy-paste commands so another maintainer can reproduce your test steps.
- Prefer: make, go test ./... (or a targeted test, e.g. go test ./cmd/samples/recipes/helloworld/)
- For cmd/samples: make then e.g. ./bin/helloworld -m worker and ./bin/helloworld -m trigger
- For new_samples: cd new_samples/<sample> && go run . and Cadence CLI commands as in the sample README
- Good: Full commands reviewers can copy-paste to verify
- Bad: "Tested locally" or "See tests" -->
**How did you test it?**


<!-- Often N/A for sample-only or doc-only changes. Call out when relevant:
- Dependency upgrades (e.g. cadence-client version)
- Behavior changes that could affect someone copying the sample
- Build or config changes
- If truly N/A, you can mark it as such -->
**Potential risks**


<!-- Optional for this repo. Use when the change is user-facing (e.g. new sample, notable README change).
- Can be N/A for internal refactors, tiny fixes, or incremental work -->
**Release notes**


<!-- Did you update the main README, a cmd/samples README, or a new_samples README (including generator-generated READMEs)?
- Any links to cadence or cadence-docs that need updating?
- Only mark N/A if you're certain no docs are affected -->
**Documentation Changes**


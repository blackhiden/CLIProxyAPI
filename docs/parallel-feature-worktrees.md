# Parallel Feature Worktrees

Use isolated feature branches when roadmap work has genuinely independent
workstreams. This keeps files separate; it does not create independent agents,
databases, services, or network resources.

## Planning and task selection

1. Read the roadmap, repository guidance, architecture, relevant specifications,
   source, tests, and Git state before planning.
2. Split only low-coupling work into separate tasks. Prefer a sequential task
   when branches would overlap on APIs, shared components, schema, or generated
   outputs.
3. Agree on one shared `main` commit. Record per-task scope, files, dependencies,
   test commands, and integration risks before creating branches.
4. Obtain approval before creating worktrees or changing application code.

## Safe worktree setup

The helper `scripts/parallel-feature-worktrees.sh` creates one branch and
worktree per task suffix from a specified base commit. Paths and branches that
already exist cause an error; the script never removes worktrees, merges
branches, deletes branches, or rewrites history.

```bash
scripts/parallel-feature-worktrees.sh \
  /path/to/CLIProxyAPI \
  <shared-main-commit> \
  /path/to/worktrees \
  feature-a feature-b
```

This creates branches `feature/feature-a` and `feature/feature-b` at separate
directories under the chosen parent. Before assigning work, create an
`AGENT_INSTRUCTIONS.md` and `INITIAL_PROMPT.md` in each worktree with the
approved task boundaries and verification requirements. Agents must only edit
their assigned worktree. Initialize submodules at the revision required by the
approved integration baseline; do not update unrelated submodule pointers.

## Isolation and shared resources

Use distinct ports and temporary directories. Never reuse or overwrite another
task's environment, `.env`, generated files, database, or running service. Do
not run commands that clean/reset other worktrees. Keep dependencies and shared
API changes sequential unless the contract is fixed and independently
testable.

Generated root assets shared by workstreams should be produced in a designated
integration/documentation worktree from the reviewed frontend submodule commit;
do not write outside an assigned worktree.

## Validation and review

Each branch must run its focused tests, lint/type checks as appropriate, and
build. Inspect `git status`, `git diff --check`, and the complete diff; record
the exact commands, outcomes, commit ID, limitations, and changed files.
Integration review should identify overlapping files, contract mismatches,
submodule pointer changes, generated asset freshness, and documentation gaps.

## Merge and cleanup

Do not merge, rebase, squash, push forcibly, or delete branches without explicit
approval. After merge approval, merge reviewed branches using the repository's
normal process, resolve conflicts deliberately, and rerun integration tests.
Remove a worktree only after its changes are safely merged and it has no
uncommitted work; delete the branch only when separately authorized.

```bash
git worktree list
git worktree remove /path/to/worktree  # only after merge and explicit cleanup approval
git branch -d feature/task-name        # only when separately authorized
```

## Common failure cases

- Existing branch or destination path: choose another unique suffix; never force
  the helper to reuse it.
- Ignored or user-local planning files: do not force-add, overwrite, or copy
  them into a feature branch; proceed only from the approved implementation
  instructions and record the limitation.
- Wrong submodule revision: inspect the superproject gitlink and checkout the
  explicitly approved submodule commit.
- Generated bundle appears stale: rebuild from the matching frontend source
  revision, then inspect the artifact and commit its parent pointer intentionally.
- A task discovers scope or API changes are needed: stop and request approval
  rather than expanding the branch silently.
- Tests share ports or local configuration: rerun with isolated resources and
  retain the original user's state.
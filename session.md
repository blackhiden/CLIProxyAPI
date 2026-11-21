# Session: Upstream Sync and v8.1.1 Release

**Date:** 2026-11-22
**Status:** Git Flow release `8.1.1` finalized

## Objective

Synchronize `upstream/main` into the fork's `dev` branch, preserve all local features, and finalize
a Git Flow release into `main` per `docs/merge-upstream.md`.

## Upstream reference

- `upstream/main` = `6960763455c4ac3d4828fb1736aa5de2183ccecf` (`v8.1.0`,
  `feat(management): persist server identity and expose it to authenticated clients`).
- 47 upstream commits; 728 files changed (+45,163 / -13,029) between the pre-merge `dev` head and
  the upstream head.

## Merge results

- `upstream/main` → `dev`: merge commit `edc1cc22a`, parents `457620047` (pre-merge `dev`) and
  `696076345` (upstream). Nine conflicts were resolved by combining both sides.
- `main` → `dev`: merge commit `24b918844`. This extra merge was required because the local
  request-timeout feature existed only on `main`; releasing from `dev` without it would leave `dev`
  inconsistent with `main`. Three conflicts were resolved.
- Follow-up commit `fced4c1d9` adapted the adversarial WebRTC test to upstream's `newMediaRelay`
  constructor after the upstream removal of `newPionMediaRelay`.

## Conflict resolutions

- `.github/workflows/pr-test-build.yml`: kept both the fork's govulncheck step and upstream's
  Solaris cross-build step.
- `go.sum`: kept both testify versions and the fork's tidwall checksums.
- `config_basic.go`: kept the fork's 30s timeout, fork release/commit URLs, and tolerant
  version+commit response. Upstream's new error-response branch was not adopted because it
  conflicts with the fork's intentional tolerant response; no tests depend on it.
- `handler.go`: combined upstream's server-identity snapshot with the fork's actor audit log and
  configurable `max-failed-attempts`/`ban-duration` settings.
- `config_load.go`: adopted upstream's `sanitizeStartupConfig()` refactor and added
  `NormalizeLiveFlowConfig()` there so the fork-only Live Flow compatibility survives all load
  paths.
- `config_v8.go`: kept both `pruneV8APIKeyNames()` and upstream's
  `removeConflictingConfigPaths()`.
- `config_yaml.go`: kept upstream's locked save path and restored `SanitizeAPIKeyNames()` before
  rendering.
- Import-only conflicts: kept upstream imports required by new code.
- `conductor_execution.go`: used upstream's payload builder while preserving the Antigravity
  thinking-model output budget through `verificationMaxOutputTokens()`.
- `CHANGELOG.md`/`session.md`: kept both histories. `static/management.html`: took the `main`
  bundle to match the preserved frontend pointer.

## Preservation verification (all pass)

- RESP error before close, TCP keep-alive, 30s version-check timeout, `oauthProxyOverride`,
  `access.api-key-names` mapping and pruning, Live Flow config/server/route, fork compose image,
  request-timeout config paths, and served bundle markers (`Request Timeout`, `do_not_use_proxy`,
  `oauth-result-modal`) are present.
- Submodule pointers: `frontend` `7ecfb8a9`, `cpauk` `bfc1f5a1`.

## Validation

- `go build -o /tmp/cpa-merge-check ./cmd/server`: pass.
- Targeted Go suites (config, watcher diff, SDK handlers/auth, management handlers, internal API,
  Live Flow media): all pass.
- Frontend targeted tests: 80 pass, 0 fail; `bun run type-check`: pass.
- Full `go test ./... -count=1 -p 4`: fails only in watcher/service tests with inotify `no space
  left on device`; the same failures reproduce on the pre-merge `main` baseline, so they are
  environmental rather than merge regressions.

## Release

- Previous latest tag: `v8.1.0`; next version derived automatically: `8.1.1` (git-flow
  `versiontag` prefix `v` produces `v8.1.1`).
- Release branch: `release/8.1.1`, started from `dev`.
- `CHANGELOG.md` gained a `[v8.1.1]` section; `[Unreleased]` is empty.

## Known issues / discrepancies

- Upstream's `config_basic.go` error-response branch was intentionally not adopted because it
  conflicts with the fork's tolerant version-check contract.
- Full-suite watcher/service failures are environmental (inotify ENOSPC) and pre-existing.

---


# Session: Fix False Update Available Notification

**Date:** 2026-11-21
**Status:** Completed

## Objective

Correct the management UI false "Update Available" modal when the running fork is ahead of the
upstream release (current `v8.0.82`, upstream `v8.0.60`, upstream commit `adf043f`).

## Changes

- Frontend submodule: shared semantic-version comparison in `src/utils/updateNotification.ts`;
  root shell and System page now gate the modal on upstream version being newer, not on commit
  inequality.
- System page main-repository quick link now uses the shared `blackhiden/CLIProxyAPI` URL.
- Added regression tests for ahead/equal/behind, multi-digit segments, and missing/malformed
  versions.
- Refreshed the generated `static/management.html` from the rebuilt frontend bundle.
- Updated root and frontend changelogs/session notes plus frontend architecture/spec docs.

## Validation

- `bun test tests/updateNotification.test.ts`: pass (9 tests).
- Frontend full suite: 2393 pass, 1 todo, 18 pre-existing failures documented in `frontend/CHANGELOG.md`.
- `bun run build`: pass.
- `go build -o /tmp/cpa-verify-build ./cmd/server`: pass.

---

# Session: Configurable Request Timeout

**Date:** 2026-11-21
**Status:** Merged into `main`; feature branches and worktrees retained pending explicit cleanup authorization

## Integration into `main`

All five feature branches were merged sequentially into `main` with
`--no-ff`, in dependency order: config (`1298bddf6`), runtime (`146f9ae15`),
frontend (`49c26fc41`), docs (`5b0467d03`), and integration (`c088d3d63`).
No conflicts occurred. Final `main` commit before this session-note update:
`c088d3d63`.

Final regression: full `go test ./... -count=1 -p 4` passed (exit 0), server
build passed, `git diff --check` passed, frontend targeted tests passed
(71 pass, 0 fail), and frontend type-check passed. The frontend gitlink is
`69bbb190` and `static/management.html` contains the new timeout labels.

`ROADMAP.md` was not changed: the root roadmap states that completed work is
recorded in `CHANGELOG.md`, which the docs branch updated. `docs/ROADMAP.md`
exists only on `dev`, not on the `main` baseline.


## Objective

Implement the roadmap feature "Configurable Request Timeout" with the accepted
recommendations: a soft long-request budget, an opt-in hard deadline, a dedicated
management UI field, and per-provider overrides.

## Baseline and worktrees

All branches originate from `main` commit
`10bb5ac65deca62b56f0618d3a9e239402ccd3e6`.

| Task | Branch | Worktree | Commit |
| ---- | ------ | -------- | ------ |
| Config contract | `feature/request-timeout-config` | `/home/denny/Project/cpa/request-timeout-config` | `a23e4c161` |
| Runtime behavior | `feature/request-timeout-runtime` | `/home/denny/Project/cpa/request-timeout-runtime` | `af7a85346` |
| Frontend UI | `feature/request-timeout-frontend` | `/home/denny/Project/cpa/request-timeout-frontend` | `5663a5362` |
| Frontend submodule | `feature/request-timeout-ui` | `frontend/` inside the frontend worktree | `69bbb190` |
| Docs/preservation | `feature/request-timeout-docs` | `/home/denny/Project/cpa/request-timeout-docs` | `d7d3b7bdd` |
| Integration | `feature/request-timeout-integration` | `/home/denny/Project/cpa/request-timeout-integration` | this integration commit |

## Implemented contract

- `requests.request-timeout` — soft budget; enables pre-first-token SSE and
  non-streaming keepalives, derives `min(budget/3, 30s)` with a 1s floor when no
  explicit interval is set, and logs one non-fatal budget-exceeded event. It
  never cancels a request.
- `requests.hard-request-deadline` — opt-in deadline for non-streaming execution
  and streaming bootstrap only; never cancels an established stream.
- Per-provider overrides under `upstream.<provider>.request-timeout` and
  `upstream.<provider>.hard-request-deadline` for codex, claude, xai,
  antigravity, and devin. Positive provider values win over global values.
- Duration parser accepts Go duration strings, bare integer seconds, and
  disabled spellings; invalid/negative values yield 0 and overflow saturates.
- v8 path mappings, hot-reload diff entries, config template, changelog,
  preservation item 14 / section 3.14, and the explicit timeout-policy exception
  are included.

## Validation

All commands passed in the integration worktree:

```bash
gofmt -l internal/config internal/watcher/diff sdk/api/handlers/*.go
go test ./internal/config ./internal/watcher/diff ./sdk/api/handlers/... ./internal/api/handlers/management -count=1
go build -o /tmp/cliproxyapi-request-timeout-check ./cmd/server && rm /tmp/cliproxyapi-request-timeout-check
git diff --check
cd frontend && bun test tests/configFieldParity.test.ts tests/visualConfigAdditions.test.ts && bun run type-check
```

Frontend targeted tests: 71 pass, 0 fail. Go packages: all pass. The full
`go test ./... -count=1 -p 4` suite also passed (exit 0). Server build: pass.
`static/management.html` was rebuilt from frontend commit `69bbb190` and
contains the new labels.

## Known limitations

- The streaming hard deadline bounds only the pre-first-token bootstrap; it does
  not cancel an established stream.
- Keepalive ticker creation happens before provider selection, so the derived
  keepalive interval uses the global budget. Per-provider overrides affect the
  budget diagnostic and the non-streaming hard deadline.
- External Nginx/Cloudflare/client read timeouts are outside this process and
  must be raised by operators.
- `docs/ROADMAP.md` and `docs/roadmap-worktree.md` exist on `dev` but not on the
  `main` baseline, so they were not modified in the feature branches.

## Review and merge instructions

Do not merge until explicitly approved. Suggested order:

```bash
git checkout dev
git merge --no-ff feature/request-timeout-config
git merge --no-ff feature/request-timeout-runtime
git merge --no-ff feature/request-timeout-frontend
git merge --no-ff feature/request-timeout-docs
git merge --no-ff feature/request-timeout-integration
```

After merging, rerun the validation commands above and confirm the frontend
gitlink points at a commit containing `69bbb190`.

## Cleanup instructions (only after explicit approval)

```bash
git worktree remove /home/denny/Project/cpa/request-timeout-config
git worktree remove /home/denny/Project/cpa/request-timeout-runtime
git worktree remove /home/denny/Project/cpa/request-timeout-frontend
git worktree remove /home/denny/Project/cpa/request-timeout-docs
git worktree remove /home/denny/Project/cpa/request-timeout-integration
git branch -d feature/request-timeout-config feature/request-timeout-runtime feature/request-timeout-frontend feature/request-timeout-docs feature/request-timeout-integration
```

---


# Session: Upstream Sync and v8.0.82 Release

**Date:** 2026-11-21
**Status:** `upstream/main` merged into `dev`; Git Flow release `8.0.82` finalized

## Objective

Synchronize `upstream/main` (`router-for-me/CLIProxyAPI`) into the fork's `dev`
branch, preserve all fork-specific features, then finalize a Git Flow release
into `main` per `docs/merge-upstream.md`.

## Upstream reference

- `upstream/main` = `f5c5681a8361c9da53e93c60b6507816eee7468b`
  (`fix(auth): capture per-attempt access token on streaming failure results`).
- 28 upstream commits merged; 96 files changed (+13,899 / -357).

## Merge result

- Merge commit `1cb1c77c92f7693086c2d81f37541df87b88f97d` on `dev`, parents
  `8681371ff` (local Docker apt-cache commit) and `f5c5681a8` (upstream).
- No conflicts.
- Follow-up commit `5e5d446d2` formatted six upstream Go files with `gofmt`;
  no repository-wide reformat was performed.
- Submodule pointers unchanged: `cpauk` `bfc1f5a1`, `frontend` `038f6a78`.

## Preservation verification (all pass)

- RESP error before close: 2 hits in `internal/api/redis_queue_protocol.go`.
- TCP keep-alive: 3 hits in `internal/api/protocol_multiplexer.go`.
- Version-check timeout 30s: present in `config_basic.go`.
- `oauthProxyOverride` helper and active v8 proxy path are present.
- `access.api-key-names` v8 mapping + `pruneV8APIKeyNames`: present.
- Live Flow: 5 `internal/api/flowviz*.go` files, `CPAForkLiveFlowEnabled` in
  config, server, and `/live-flow/ws` route registration.
- Fork compose image `ghcr.io/blackhiden/cliproxyapi:latest` in both compose files.
- Served bundle contains `Management Center Info`, `do_not_use_proxy`,
  `oauth-result-modal`, and `Test Model`.

## Validation

- `go build -o /tmp/cpa-verify-build ./cmd/server`: pass.
- Full `go test ./... -count=1 -p 4`: exit 0, no failures.
- `gofmt` check on changed Go files: clean after the formatting commit.

## Release

- Previous latest tag: `v8.0.81`; next version derived automatically: `8.0.82`
  (git-flow `versiontag` prefix `v` produces tag `v8.0.82`).
- Release branch: `release/8.0.82`, started from `dev`.
- `CHANGELOG.md` gained a `[v8.0.82]` section; `[Unreleased]` is empty.
- `docs/MERGE-PRESERVATION.md` corrected the stale fork image reference to
  `ghcr.io/blackhiden/cliproxyapi:latest`.

## Known issues / discrepancies

- `docs/MERGE-PRESERVATION.md` previously documented
  `ghcr.io/arsydoni4326-alt/cliproxyapi:latest`; the actual fork image is
  `ghcr.io/blackhiden/cliproxyapi:latest`. The current contract was corrected;
  the historical changelog entry was left unchanged.
- No known test failures after this merge.

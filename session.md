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

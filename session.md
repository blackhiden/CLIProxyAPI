# Session: Upstream Sync and v8.0.81 Release

**Date:** 2026-11-20
**Status:** `upstream/main` merged into `dev`; Git Flow release `8.0.81` prepared

## Objective

Synchronize `upstream/main` (`router-for-me/CLIProxyAPI`) into the fork's `dev`
branch, preserve all fork-specific features, then finalize a Git Flow release
into `main` per `docs/merge-upstream.md`.

## Upstream reference

- `upstream/main` = `6faa8d20b2673404cff70a2f2d9537ae0d0e684d`
  (`fix(codex): strip uniqueItems keyword from tool parameter schemas`).
- 16 upstream commits merged; 91 files changed (+6083 / -325).

## Merge result

- Merge commit `d64c844340e79bade334a15945362df84fcf5bfc` on `dev`.
- Conflicts (2):
  - `internal/config/config_v8.go`: kept the fork's `pruneV8APIKeyNames` call and
    upstream's `commented []ArchivedV8Section` tracking in
    `normalizeConfigLayout`; both behaviors are required.
  - `internal/runtime/executor/codex_websockets_executor.go`: kept upstream's
    `sdktranslator` import, required by the new session-instruction marker
    validation calls.
- The local `cpauk` submodule pointer change (`2647cf9e9` → `3968abcbb`) was
  preserved as an unstaged local change and was not included in the release.

## Preservation verification (all pass)

- RESP error before close: 2 hits in `internal/api/redis_queue_protocol.go`.
- TCP keep-alive: 3 hits in `internal/api/protocol_multiplexer.go`.
- Version-check timeout 30s: present in `config_basic.go`.
- `oauthProxyOverride` helper and active v8 `oauthSessionProxyURLFromContext`
  path are present.
- `access.api-key-names` v8 mapping + `pruneV8APIKeyNames`: present.
- Live Flow: 5 `internal/api/flowviz*.go` files, `CPAForkLiveFlowEnabled` in
  config, server, and route registration.
- Fork compose image `ghcr.io/blackhiden/cliproxyapi:latest` in both compose files.
- Served bundle contains `Management Center Info`, `do_not_use_proxy`, and
  `oauth-result-modal`.

## Validation

- `go build -o /tmp/cpa-merge-build ./cmd/server`: pass.
- Full `go test ./... -count=1` with bounded parallelism: 101 packages ok, 31
  with no test files, 4 failing packages:
  - `cmd/server`, `internal/watcher`, `sdk/cliproxy`: filesystem watcher tests
    fail with `no space left on device` because the host has 128 inotify
    instances and 98 are already held by desktop processes. The same failures
    reproduce on a clean pre-merge `dev` worktree, so they are environmental.
  - `internal/api`: `TestModelsWithClientVersionReturnsCodexCatalog` is the
    previously documented upstream catalog defect (`gpt-5.5` has empty
    `service_tiers`).
- Conflict-area packages pass: `internal/config`, `internal/runtime/executor`,
  and `internal/api/handlers/management`.
- Fork-specific tests pass: v8 API-key-name persistence and validation, Live
  Flow persistence, Redis protocol, and OAuth proxy override.
- The two resolved Go files were formatted with `gofmt`; no repo-wide reformat
  was performed.

## Release

- Previous latest tag: `v8.0.80`; next version derived automatically: `8.0.81`
  (git-flow `versiontag` prefix `v` produces tag `v8.0.81`).
- Release branch: `release/8.0.81`, started from `dev`.
- `CHANGELOG.md` gained a `[v8.0.81]` section; `[Unreleased]` is empty.

## Known issues

- The upstream catalog test failure will also affect `main` after the release.
- Watcher-related tests require a higher inotify instance limit than this host
  currently provides; this is unrelated to the merge.

# Session: Upstream Sync and v8.0.80 Release

**Date:** 2026-11-18
**Status:** `upstream/main` merged into `dev`; Git Flow release `8.0.80` prepared

## Objective

Synchronize `upstream/main` (`router-for-me/CLIProxyAPI`) into the fork's `dev`
branch, preserve all fork-specific features, then finalize a Git Flow release
into `main` per `docs/merge-upstream.md`.

## Upstream reference

- `upstream/main` = `a5e8f69e56309e1ef6d6872e667da81fdfbce2ce`
  (`feat(codex): add GPT-6.2 models and update CLI version`).
- 17 upstream commits merged; 105 files changed (+9182 / -838).

## Merge result

- Merge commit `9cfc242baca0b68b0a4f6b9b6812fd16235830ce` on `dev`.
- Conflicts (2, both trivial import conflicts):
  - `internal/runtime/executor/helps/usage_helpers.go`
  - `internal/runtime/executor/helps/usage_helpers_test.go`
  - Resolution: kept upstream's `sdk/pluginapi` import, required by the new
    `ApplyPluginExecutorAttribution` path. No fork behavior was involved.
- No submodule pointer, `static/management.html`, compose, or `config.yaml`
  changes were introduced by the merge.

## Preservation verification (all pass)

- RESP error before close: 2 hits in `internal/api/redis_queue_protocol.go`.
- TCP keep-alive: 3 hits in `internal/api/protocol_multiplexer.go`.
- Version-check timeout 30s: present in `config_basic.go`.
- `oauthProxyOverride`: helper present; the active v8 path uses it through
  `oauthSessionProxyURLFromContext` (`auth_files_v8.go`). The legacy
  `/v0/management` handlers remain single-argument, unchanged before and after
  the merge (deprecated endpoints).
- `access.api-key-names` v8 mapping + `pruneV8APIKeyNames`: present.
- Live Flow: 5 `internal/api/flowviz*.go` files, `CPAForkLiveFlowEnabled` in
  config and `server.go`.
- Fork compose image `ghcr.io/blackhiden/cliproxyapi:latest` in both compose files.
- Served bundle contains `Management Center Info`, `do_not_use_proxy`, and
  `oauth-result-modal`.

## Validation

- `go build -o /tmp/cpa-merge-build ./cmd/server`: pass.
- `go test ./... -count=1`: 103 packages ok, 2 failures:
  - `TestModelsWithClientVersionReturnsCodexCatalog` (`internal/api`) — reproduced
    identically on a pristine `upstream/main` worktree, so it is an upstream
    regression (catalog ships empty `service_tiers` for `gpt-5.5` while the test
    expects one priority tier). Not caused by this merge; left untouched.
  - `TestHubExchange_ManyToManyFullDuplexMultiplexing` (`internal/wstunnel`) —
    flaky timing assertion; passes on rerun and on repeated package runs.
    `internal/wstunnel` was not touched by upstream.
- Fork-specific tests pass: flowviz/redis-protocol/oauth-proxy (`internal/api`),
  v8 API-key-name and Live Flow normalization (`internal/config`), and
  `TestOAuthProxyOverride`/`TestConfigV8SaveAPIKeyNames` (management handlers).
- `gofmt -l` reports 321 pre-existing files (trailing blank lines before closing
  parens/braces) that predate this merge; the two resolved files are clean.
  Repo-wide reformatting was intentionally not performed (scope discipline).

## Release

- Previous latest tag: `v8.0.79`; next version derived automatically: `8.0.80`
  (git-flow `versiontag` prefix `v` produces tag `v8.0.80`).
- Release branch: `release/8.0.80`, started from `dev`.
- `CHANGELOG.md` gained a `[v8.0.80]` section; `[Unreleased]` is now empty.

## Known issues

- Upstream `TestModelsWithClientVersionReturnsCodexCatalog` failure (see above)
  will also affect `main` after the release; it is an upstream defect.
- `internal/wstunnel` hub-exchange test is timing-flaky under load.

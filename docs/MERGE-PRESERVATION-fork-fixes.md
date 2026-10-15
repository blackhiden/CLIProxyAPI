# Merge Preservation — Fork-Specific Fixes

> **Purpose of this document:** these changes are *local* to the
> `arsydoni4326-alt/CLIProxyAPI` fork and are **not** part of upstream
> `router-for-me/CLIProxyAPI`. When merging upstream into this fork, a naive
> merge can silently drop or revert them — either because upstream modified the
> same lines (conflict resolved the wrong way) or because the change lives in a
> generated/submodule file that gets overwritten. This document tells you (and
> any AI agent performing the merge) exactly what to preserve, where it lives,
> how to verify it survived, and how to restore it if it was lost.
>
> **Rule for merges:** when a conflict touches any file listed in §2, resolve in
> favor of the fork behavior described in §3 unless upstream explicitly
> implements the same behavior — in that case, prefer upstream and delete the
> fork-specific patch, then update this document.

- **Documented:** 2026-09-12
- **Applies to:** fork `develop` branch, post-`v7.2.158-arsydoni4326-alt`
- **Related changelog:** `CHANGELOG.md` section `[Unreleased]`

---

## 1. Change Inventory (quick reference)

| # | Change | File(s) | Type | Loss mode |
|---|--------|---------|------|-----------|
| 1 | RESP error before close when management disabled | `internal/api/redis_queue_protocol.go` | Go source | conflict resolution reverts to abrupt `return` |
| 2 | TCP keep-alive on accepted connections | `internal/api/protocol_multiplexer.go` | Go source | conflict resolution drops `enableAcceptedKeepAlive(conn)` call / helper |
| 3 | Version-check HTTP client timeout 30s | `internal/api/handlers/management/config_basic.go` | Go source | conflict resolution reverts to `10 * time.Second` |
| 4 | Updated redis-protocol integration test | `internal/api/redis_queue_protocol_integration_test.go` | Go test | conflict restores old abrupt-close assertions |
| 5 | Locale files restructured to top-level sections | `frontend/src/i18n/locales/{en,zh-CN,zh-TW,ru}.json` | Submodule content | submodule pointer overwrite / upstream submodule merge |
| 6 | Rebuilt static management bundle | `static/management.html` | Generated artifact | stale artifact shipped by upstream replaces it |
| 7 | Compose defaults to the fork image | `docker-compose.yml`, `docker-compose.cluster.yml` | Deployment config | upstream merge restores `eceasy/cli-proxy-api:latest`, so deployments silently lose items 1–6 |
| 8 | OAuth "Do Not Use Proxy" toggle (default checked) | `internal/api/handlers/management/auth_files_provider_oauth.go`, `internal/api/handlers/management/auth_files.go`, `internal/auth/kimi/kimi.go`, `internal/api/handlers/management/oauth_no_proxy_test.go`, `frontend/src/pages/OAuthPage.tsx`, `frontend/src/services/api/oauth.ts`, `frontend/src/pages/OAuthPage.module.scss`, `frontend/src/i18n/locales/{en,zh-CN,zh-TW,ru}.json`, `static/management.html` | Go source + Submodule content + Generated artifact | conflict resolution reverts handlers to `New*Auth(cfg)`; submodule pointer overwrite; stale upstream bundle replaces the checkbox |
| 9 | OAuth results shown as a manually-dismissed modal (not toasts) | `frontend/src/components/common/OAuthResultModal.tsx`, `frontend/src/pages/OAuthPage.tsx`, `frontend/src/styles/components.scss`, `frontend/tests/oauthResultModal.test.ts` | Submodule content | conflict resolution restores `showNotification(...)` calls; submodule pointer overwrite; stale upstream bundle replaces the modal |
| 10 | **Live Flow feature (persistent)** | `internal/api/flowviz.go`, `internal/api/flowviz_test.go`, `internal/api/flowviz_bench_test.go`, `internal/api/flowviz_conversation_recovery_test.go`, `internal/api/server.go`, `internal/config/config.go`, `internal/config/config_normalization.go`, `internal/config/config_load.go`, `config.yaml`, `config.example.yaml` | Go source + Config | conflict resolution removes feature; upstream removes flowviz files; config key renamed/removed |
| 11 | **cpauk full own-event request logs (persistent)** | `cpauk/internal/api/key_viewer_request_events.go`, `cpauk/internal/api/test/key_viewer_request_events_test.go`, `cpauk/docs/MERGE-PRESERVATION.md`, `cpauk/docs/{SPECIFICATION,CONFIGURATION}.md` | Submodule code + tests + docs | submodule pointer update or upstream merge restores request-log redaction |

---

## 2. How each change can be lost in a merge

### 2.1 Direct conflict (Go sources, items 1–4)

Upstream owns these files too. If upstream edits
`redis_queue_protocol.go`, `protocol_multiplexer.go`, or `config_basic.go`
near the patched lines, `git merge` raises conflicts. The dangerous failure
mode is **not** the conflict itself but resolving it with `--theirs`/checkout of
the upstream side "to take upstream cleanly".

### 2.2 Submodule pointer (item 5)

The `frontend/` directory is a git submodule pointing at
`arsydoni4326-alt/Cli-Proxy-API-Management-Center`. The locale fix exists as
**commits inside the submodule repo**, not in this repo. A merge that moves the
submodule pointer (e.g. upstream bumps it, or a merge tool resets `frontend` to
upstream's recorded SHA) silently replaces the whole frontend tree, including
the locale fix.

### 2.3 Generated artifact (item 6)

`static/management.html` is checked in but is a build product of the frontend.
Upstream ships its own `management.html`; a merge that takes upstream's copy
reintroduces the stale translations (raw `system_info.*` keys rendered in the
UI) even though the source fix in the submodule is intact.

### 2.4 Cross-cutting feature (item 8 — OAuth "Do Not Use Proxy")

Item 8 spans all three loss modes at once:

1. **Go handlers**: upstream may refactor `Request*Token` in
   `internal/api/handlers/management/auth_files_provider_oauth.go` or the
   `newCodexOAuthService` factory in `internal/api/handlers/management/auth_files.go`.
   A conflict resolved with the upstream side silently restores
   `claude.NewClaudeAuth(h.cfg)` / `newCodexOAuthService(h.cfg)` style calls, i.e.
   OAuth logins always use the configured proxy again.
2. **Submodule**: the checkbox lives in
   `frontend/src/pages/OAuthPage.tsx` + `frontend/src/services/api/oauth.ts` inside
   `arsydoni4326-alt/Cli-Proxy-API-Management-Center`. A merge that moves the
   submodule pointer replaces the whole frontend tree, dropping the checkbox.
3. **Generated artifact**: upstream's `static/management.html` does not contain the
   checkbox, so a merge taking upstream's bundle removes the feature from the UI
   even if the sources survived.

---

## 3. The exact behaviors to preserve (with verification)

### 3.1 RESP error before close (item 1)

**File:** `internal/api/redis_queue_protocol.go`, in
`(*Server).handleRedisConnection`, top of the `for` loop.

**Required code:**

```go
for {
    if !s.managementRoutesEnabled.Load() {
        // Reply with a proper Redis error before closing so clients (e.g.
        // cpa-usage-keeper) get a meaningful message instead of a TCP RST
        // ("connection reset by peer") on their pending AUTH/SUBSCRIBE read.
        _ = writeRedisError(writer, "ERR remote management disabled")
        _ = writer.Flush()
        return
    }
    // ...
```

**Anti-pattern to reject:** a bare

```go
for {
    if !s.managementRoutesEnabled.Load() {
        return
    }
```

**Verify after merge:**

```bash
grep -n 'ERR remote management disabled' internal/api/redis_queue_protocol.go
go test ./internal/api/ -run TestRedisProtocol_ManagementDisabled_RejectsConnection -count=1
```

The test must pass. If the test now asserts "connection closed without any
response", the old test was restored by the merge — reapply item 4.

### 3.2 cpauk full own-event request logs (item 11)

**Scope:** the `cpauk` submodule. Its API-key viewer route
`GET /api/v1/key-events/:id/request-log` must return every CPA log section for
the event owner when `CPA_REQUEST_LOG_ACCESS_ENABLED=true`. It must not restore
`redactRequestLogForKeyViewer` or another filtering step after ownership
verification.

**Required security boundary:** authenticated viewers receive full logs only
for events owned by their API key. Cross-key access stays `404`; disabled
request-log access stays `403`.

**Verify after a root or submodule merge:**

```bash
cd cpauk
grep -n 'FORK-ONLY: The owning API-key viewer receives every request-log section' internal/api/key_viewer_request_events.go
go test ./internal/api/test/ -run 'TestAPIKeyViewerRequestLogFullContentPersistence|TestAPIKeyViewerRequestLogRejectsCrossKeyAccess|TestAPIKeyViewerRequestLogRejectsDisabledAccess' -count=1
```

The test compares every returned log section with its original content. A
restored redaction policy must fail this test. The submodule-local guide is
[`cpauk/docs/MERGE-PRESERVATION.md`](../cpauk/docs/MERGE-PRESERVATION.md).

### 3.3 TCP keep-alive on accepted connections (item 2)

**File:** `internal/api/protocol_multiplexer.go`.

**Required code:**

```go
// acceptedKeepAliveConfig keeps NAT/overlay middleboxes from silently dropping
// idle long-lived connections (redis usage subscriptions, streaming responses).
var acceptedKeepAliveConfig = net.KeepAliveConfig{
    Enable:   true,
    Idle:     30 * time.Second,
    Interval: 15 * time.Second,
    Count:    6,
}

// enableAcceptedKeepAlive turns on TCP keep-alive probes for an accepted
// connection; failures are non-fatal and only logged.
func enableAcceptedKeepAlive(conn net.Conn) {
    if tcp, ok := conn.(*net.TCPConn); ok {
        if err := tcp.SetKeepAliveConfig(acceptedKeepAliveConfig); err != nil {
            log.Debugf("failed to enable TCP keep-alive: %v", err)
        }
    }
}
```

and, as the **first statement** of `routeMuxConnection`:

```go
func (s *Server) routeMuxConnection(conn net.Conn, httpListener *muxListener) {
    enableAcceptedKeepAlive(conn)
    // ...
```

**Anti-patterns to reject:** the helper existing but the call missing from
`routeMuxConnection` (silent dead code after a bad conflict resolution), or the
helper deleted entirely.

**Verify after merge:**

```bash
grep -n 'enableAcceptedKeepAlive' internal/api/protocol_multiplexer.go   # expect 3 hits: doc comment, definition, call in routeMuxConnection
go build -o /tmp/test-output ./cmd/server && rm /tmp/test-output
```

### 3.3 Version-check timeout 30s (item 3)

**File:** `internal/api/handlers/management/config_basic.go`, in
`(*Handler).GetLatestVersion`.

**Required code:**

```go
client := &http.Client{Timeout: 30 * time.Second}
```

**Anti-pattern to reject:** `Timeout: 10 * time.Second`.

**Verify after merge:**

```bash
grep -n 'http.Client{Timeout: 30' internal/api/handlers/management/config_basic.go
```

### 3.4 Integration test contract (item 4)

**File:** `internal/api/redis_queue_protocol_integration_test.go`,
`TestRedisProtocol_ManagementDisabled_RejectsConnection`.

**Required behavior:** the test reads a RESP error line and expects exactly
`ERR remote management disabled`, then asserts the connection is closed. It
must **not** assert "connection closed without any response".

**Verify after merge:**

```bash
grep -n 'remote management disabled' internal/api/redis_queue_protocol_integration_test.go
```

### 3.5 Frontend locale structure (item 5 — submodule)

**Where the fix lives:** inside the submodule
`arsydoni4326-alt/Cli-Proxy-API-Management-Center`, in
`src/i18n/locales/{en,zh-CN,zh-TW,ru}.json`.

**Required structure:**

- These sections exist at the **top level** of each locale JSON:
  `system_info`, `footer`, `notification`, `language`, `theme`, `sidebar`,
  `plugin_management`, `plugin_resource`, `plugin_store`, `providersPage`,
  `quota_management`.
- The `config_management` object contains only config-panel-specific keys
  (`actions`, `diff`, `editor_placeholder`, `meta_*`, `mode`, `reload`, `save`,
  `search_*`, `status_*`, `title`, `validation_blocked_short`, `visual`,
  `visual_mode_*`) — it must **not** contain the sections listed above.
- `en.json` and `ru.json` contain the full `system_info` block (26+ keys incl.
  `title`, `about_title`, `models_*`, `version_*`, `quick_links_*`, `link_*`,
  `clear_login_*`).
- Locale files collectively cover every `t('...')` key referenced by `src/`
  (0 missing keys per repo-wide scan).

**Why:** components call `useTranslation()` without a keyPrefix and use
`t('system_info.title')`, `t('footer.version')`, `t('notification.*')` etc.
Nesting them under `config_management.*` renders raw keys in the UI. Do **not**
"fix" a merge conflict by moving component keys under `config_management.*` —
the component side is upstream-compatible and the locale side is the fork fix.

**Verify after merge (run inside `frontend/`):**

```bash
python3 - <<'EOF'
import json
d = json.load(open('src/i18n/locales/en.json'))
print('top-level system_info present:', 'system_info' in d)
print('system_info keys:', len(d.get('system_info', {})))
print('footer top-level:', 'footer' in d)
print('notification top-level:', 'notification' in d)
print('config_management has no system_info:', 'system_info' not in d['config_management'])
EOF
```

Expected: `system_info` present with >= 26 keys, `footer`/`notification`
top-level present, and `config_management` without `system_info`.

Also confirm the submodule pointer still references a SHA that contains the
fix (not upstream's frontend SHA):

```bash
git -C frontend log --oneline -3
git submodule status frontend
```
### 3.6 Static management bundle (item 6)

**File:** `static/management.html` (generated; refreshed from
`frontend/dist/index.html` after a frontend build).

**Required content:** the served bundle must embed the top-level
`system_info` translations for all four languages.

**Verify after merge:**

```bash
grep -c '管理中心信息' static/management.html             # zh-CN present
grep -c 'Management Center Info' static/management.html  # en present
```

If these are absent, the merge brought in a stale/upstream bundle. Rebuild:

```bash
# Inside frontend/: bun install && bun run build, then from repo root:
cp frontend/dist/index.html static/management.html
```

(If Bun is unavailable, the build can be driven via the Vite JS API with
`configFile: false`; see the `session.md` 2026-09-12 entry for the working
recipe. Note `frontend/node_modules/` may be root-owned; Vite needs write
access or the JS-API approach.)

### 3.7 Compose defaults to the fork image (item 7)

**Files:** `docker-compose.yml`, `docker-compose.cluster.yml`.

The upstream repository ships `eceasy/cli-proxy-api:latest` as the default CPA image.
Because these compose files also set `pull_policy: always`, a deployment started with
`docker compose up -d` runs the upstream binary and silently drops every fork fix —
which is how the "connection reset by peer" report came back after items 1–6 were
already in the source tree.

**Required content:** the default image is the fork's own build, with the upstream
image still selectable through the environment variable:

```yaml
image: ${CLI_PROXY_IMAGE:-ghcr.io/arsydoni4326-alt/cliproxyapi:latest}
```

**Verify after merge:**

```bash
grep -n 'ghcr.io/arsydoni4326-alt/cliproxyapi' docker-compose.yml docker-compose.cluster.yml
```

After changing which image a host runs, recreate the container
(`docker compose up -d --force-recreate`) instead of only restarting it.

### 3.8 OAuth "Do Not Use Proxy" toggle (item 8)

**Purpose:** the management UI OAuth page (`/management.html#/oauth`) shows a
**"Do Not Use Proxy"** checkbox, **checked by default**. Checked → every
OAuth/device-code login (Codex, Anthropic/Claude, Antigravity, Kimi, xAI/Grok)
connects directly, bypassing the configured `proxy-url` and environment proxies.
Unchecked → the login uses the configured proxy. The toggle is read per login
request, so check/uncheck takes effect on the next login **without restarting
the service**. Upstream does not have this feature — it must survive every merge.

**Files (Go side):**

- `internal/api/handlers/management/auth_files_provider_oauth.go` — helper
  `oauthProxyOverride` and its use in all five `Request*Token` handlers.
- `internal/api/handlers/management/auth_files.go` — `newCodexOAuthService`
  factory signature takes a proxy override.
- `internal/auth/kimi/kimi.go` — `NewKimiAuthWithProxyURL` constructor.
- `internal/api/handlers/management/oauth_no_proxy_test.go` — pinning test.

**Required code (Go side):**

```go
// oauthProxyOverride returns an explicit proxy override for OAuth login requests.
// When the management UI sends no_proxy=true ("Do Not Use Proxy" checked), the
// override value "direct" makes the OAuth flow bypass both the configured proxy
// and environment proxies. An empty result keeps the configured proxy behavior.
// The value is read per request, so toggling takes effect without a restart.
func oauthProxyOverride(c *gin.Context) string {
    if c == nil || c.Request == nil {
        return ""
    }
    switch strings.ToLower(strings.TrimSpace(c.Query("no_proxy"))) {
    case "1", "true", "yes", "on":
        return "direct"
    default:
        return ""
    }
}
```

and each handler must construct its auth service with the override:

```go
anthropicAuth := claude.NewClaudeAuthWithProxyURL(h.cfg, oauthProxyOverride(c))          // anthropic
openaiAuth := newCodexOAuthService(h.cfg, oauthProxyOverride(c))                          // codex
authSvc := xaiauth.NewXAIAuthWithProxyURL(h.cfg, oauthProxyOverride(c))                   // xai
kimiAuth := kimi.NewKimiAuthWithProxyURL(h.cfg, oauthProxyOverride(c))                    // kimi
// antigravity: non-nil http.Client (with no transport) when the override is set
var antigravityClient *http.Client
if oauthProxyOverride(c) != "" {
    antigravityClient = &http.Client{}
}
authSvc := antigravity.NewAntigravityAuth(h.cfg, antigravityClient)
```

The factory must keep the two-argument signature:

```go
newCodexOAuthService = func(cfg *config.Config, proxyURL string) codexOAuthService { return codex.NewCodexAuthWithProxyURL(cfg, proxyURL) }
```

**Anti-patterns to reject:** `claude.NewClaudeAuth(h.cfg)`,
`newCodexOAuthService(h.cfg)` (single-argument call or factory), or the
`oauthProxyOverride` helper deleted.

**Files (submodule side, `arsydoni4326-alt/Cli-Proxy-API-Management-Center`):**

- `src/pages/OAuthPage.tsx` — `doNotUseProxy` state initialized to `true`, the
  checkbox block above the provider cards (`styles.proxyToggle`), and
  `oauthApi.startAuth(provider, { noProxy: doNotUseProxy }, attempt.signal)`.
- `src/services/api/oauth.ts` — `startAuth` accepts the fork's options object
  (`no_proxy: true` when `noProxy` is set) **and** upstream's `AbortSignal`
  argument in either position (`optionsOrSignal instanceof AbortSignal`
  dispatch), so upstream request-cancellation and the fork proxy bypass
  coexist after the 2026-09-14 upstream merge.
- `src/pages/OAuthPage.module.scss` — `.proxyToggle*` styles.
- `src/i18n/locales/{en,zh-CN,zh-TW,ru}.json` — `auth_login.do_not_use_proxy`
  and `auth_login.do_not_use_proxy_hint` keys in all four locales.

**Verify after merge:**

```bash
# Go side
grep -n 'oauthProxyOverride' internal/api/handlers/management/auth_files_provider_oauth.go   # helper + 5 handler call sites
grep -n 'NewKimiAuthWithProxyURL' internal/auth/kimi/kimi.go
grep -n 'codex.NewCodexAuthWithProxyURL' internal/api/handlers/management/auth_files.go
go test ./internal/api/handlers/management/ -run TestOAuthProxyOverride -count=1

# Served bundle (repo root) — checkbox must be in the built artifact
grep -c 'do_not_use_proxy' static/management.html   # expect >= 1
```

If the Go-side greps fail, reapply the handler/factory changes above. If the
bundle grep fails, the merge brought in upstream's stale `management.html` —
rebuild the frontend (see §3.6) after restoring the submodule sources.

### 3.9 OAuth result modal (item 9)

**Purpose:** on the management UI OAuth page (`/management.html#/oauth`), every OAuth
process result — login success/failure, polling errors, callback validation warnings,
Devin session cancellation, and Vertex/iFlow import outcomes — is presented as a
centered **modal dialog** instead of a small transient toast notification. Toasts were
easy to miss (small, auto-swiped away after a few seconds). The modal uses the shared
`Modal` component, is localized in all four languages via existing `common.*` keys, and
**always requires manual dismissal** — it must never auto-dismiss on a timer. Upstream
does not have this feature — it must survive every merge.

**Files (frontend submodule):**

- `frontend/src/components/common/OAuthResultModal.tsx` — the modal component
  (`OAuthResult`, `OAuthResultType`, `OAuthResultModal`).
- `frontend/src/pages/OAuthPage.tsx` — `showOauthResult(...)` calls replace the former
  `showNotification(...)` calls for all OAuth process outcomes; only the
  “link copied” clipboard feedback still uses `showNotification`.
- `frontend/src/styles/components.scss` — appended `.oauth-result-modal-*` block.
- `frontend/tests/oauthResultModal.test.ts` — pinning tests (rendering, locales,
  no auto-dismiss timer in the component source).

**Required code (signature):**

```tsx
export type OAuthResultType = 'success' | 'error' | 'warning' | 'info';
export interface OAuthResult { type: OAuthResultType; message: string; }
export function OAuthResultModal({ result, onClose }: {
  result: OAuthResult | null;
  onClose: () => void;
}): JSX.Element | null;
```

**Strict result-message rule (must hold after every merge):**

For the *main OAuth process result* of every provider (built-in and plugin), the modal
shows **only** the localized `*_oauth_status_success` or `*_oauth_status_error` message —
never a concatenated runtime error, status code, or other detail, and never the
`*_oauth_status_waiting` status text:

```typescript
// Correct — i18n result key only, no interpolation/concatenation:
showOauthResult('success', getProviderTextByID(provider, 'oauth_status_success'));
showOauthResult('error', getProviderTextByID(provider, 'oauth_status_error'));
```

### 3.10 Live Flow feature (item 10) — FORK-ONLY PERSISTENT FEATURE

**Purpose:** The Live Flow feature provides real-time request flow visualization
through a WebSocket endpoint (`/live-flow/ws`) that streams per-request metadata
(no payloads or secrets) to the Management UI. This is a **fork-only persistent feature**
that must never be removed by upstream merges. The feature is namespaced under
`cpa-fork-liveflow` to prevent upstream conflicts.

**Critical files (all must survive every merge):**

- `internal/api/flowviz.go` — core hub, pub-sub, WebSocket observer (marked FORK-ONLY at top)
- `internal/api/flowviz_test.go` — unit tests for flow event handling
- `internal/api/flowviz_bench_test.go` — performance benchmarks for idle/active hub
- `internal/api/flowviz_conversation_recovery_test.go` — conversation state recovery tests
- `internal/api/server.go` — hub instantiation and `/live-flow/ws` route attachment
- `internal/config/config.go` — `CPAForkLiveFlowEnabled` field (marked FORK-ONLY)
- `internal/config/config_normalization.go` — `NormalizeLiveFlowConfig()` for backward compat
- `internal/config/config_load.go` — calls `NormalizeLiveFlowConfig()` during load
- `config.yaml` — `cpa-fork-liveflow-enabled: true` (marked FORK-ONLY)
- `config.example.yaml` — documents both new and legacy keys

**Required config keys:**

```yaml
# FORK-ONLY: CPA fork's persistent Live Flow feature
cpa-fork-liveflow-enabled: true   # New fork-specific key (preferred)
flow-visualization-enabled: false # Legacy key (deprecated, backward compat only)
```

**Required code in `internal/config/config.go`:**

```go
// FORK-ONLY: CPA fork's persistent Live Flow feature.
// This feature is fork-specific and must not be removed by upstream merges.
// See docs/MERGE-PRESERVATION-fork-fixes.md for merge policy.
//
// CPAForkLiveFlowEnabled mirrors lightweight per-request metadata...
CPAForkLiveFlowEnabled bool `yaml:"cpa-fork-liveflow-enabled" json:"cpa-fork-liveflow-enabled"`

// Deprecated: use cpa-fork-liveflow-enabled instead. Kept for backward compatibility.
FlowVisualizationEnabled bool `yaml:"flow-visualization-enabled" json:"flow-visualization-enabled"`
```

**Required code in `internal/api/server.go`:**

```go
// FORK-ONLY: CPA fork's persistent Live Flow feature.
if cfg.CPAForkLiveFlowEnabled {
    s.flowHub = newFlowHub()
    engine.Use(flowVizMiddleware(s.flowHub))
    handlers.SetFlowObserver(newFlowWSObserver(s.flowHub))
}
```

**Anti-patterns to reject (merge conflicts that would break the feature):**

1. Upstream removes `internal/api/flowviz*.go` files → **restore all four files from fork**
2. Config field renamed back to `FlowVisualizationEnabled` only → **keep both fields**
3. Server.go check reverts to `cfg.FlowVisualizationEnabled` → **use `cfg.CPAForkLiveFlowEnabled`**
4. Config files drop the fork-specific key → **restore `cpa-fork-liveflow-enabled: true`**
5. FORK-ONLY markers removed → **restore all markers**

**Verify after merge:**

```bash
# 1. Check core files exist
ls -1 internal/api/flowviz*.go | wc -l   # expect 4 files
grep -c 'FORK-ONLY.*Live Flow' internal/api/flowviz.go  # expect >= 1

# 2. Check config structure
grep -c 'CPAForkLiveFlowEnabled' internal/config/config.go  # expect >= 1
grep -c 'cpa-fork-liveflow-enabled' config.yaml  # expect >= 1

# 3. Check server integration
grep -c 'cfg.CPAForkLiveFlowEnabled' internal/api/server.go  # expect >= 1

# 4. Run tests
go test ./internal/api/ -run TestFlowHub -v
go build -o /tmp/test-cpa ./cmd/server  # must compile
```

**If any grep returns 0, the feature was lost — restore from the fork's last known-good commit.**



// Wrong — appending error details to the modal message:
showOauthResult('error', `${getProviderTextByID(provider, 'oauth_status_error')} ${err}`); // ❌
```

The inline status text rendered on the provider *card* (`oauth_status_error` + error
detail near line 813) is intentionally **not** part of this rule — that is the on-page
status line, not the modal. Callback validation, Devin cancellation, and
Vertex/iFlow import outcomes may keep their own explicit messages because they are not
the main login result.

**How it can be lost:**

1. Upstream touching `OAuthPage.tsx` can conflict with the `showOauthResult` calls;
   resolving with the upstream side silently restores `showNotification(...)` toasts.
2. A submodule pointer move replaces the whole frontend tree, dropping the component.
3. Upstream's `static/management.html` does not contain the modal, so taking
   upstream's bundle removes the feature from the UI even if the sources survived.

**Verify after merge (inside `frontend/`):**

```bash
test -f src/components/common/OAuthResultModal.tsx
bun test tests/oauthResultModal.test.ts
grep -c 'showOauthResult' src/pages/OAuthPage.tsx      # expect >= 12
grep -c 'showNotification' src/pages/OAuthPage.tsx     # expect 2 (declaration + copy link only)

# Strict result-message rule — the two main-result modal calls must be i18n-key-only
grep -n "showOauthResult('success', getProviderTextByID(provider, 'oauth_status_success'))" src/pages/OAuthPage.tsx  # expect >= 1
grep -n "showOauthResult('error', getProviderTextByID(provider, 'oauth_status_error'))" src/pages/OAuthPage.tsx      # expect >= 2 (polling + start errors)

# No interpolations in main OAuth result modal messages (cancel/callback flows excluded)
grep "showOauthResult('error'" src/pages/OAuthPage.tsx | grep -c '\${'  # expect 1 (devin cancel only)
grep "showOauthResult('success'" src/pages/OAuthPage.tsx | grep -c '\${' # expect 0

# Served bundle (repo root) — modal must be in the built artifact
grep -c 'oauth-result-modal' static/management.html    # expect >= 1
```

If the component or the `showOauthResult` calls are missing, restore them from this
section. If a main OAuth result call interpolates an error string (the
`oauth_status_error` call includes `${...}`), restore the strict i18n-key-only version
shown above. If the bundle grep fails, rebuild the frontend (see §3.6) after restoring
the submodule sources.

---

## 4. Post-merge verification checklist

Run all of these after every upstream merge; all must pass:

```bash
# 1. Behaviors present
grep -n 'ERR remote management disabled' internal/api/redis_queue_protocol.go
grep -n 'enableAcceptedKeepAlive' internal/api/protocol_multiplexer.go        # comment + helper + call
grep -n 'http.Client{Timeout: 30' internal/api/handlers/management/config_basic.go
grep -n 'oauthProxyOverride' internal/api/handlers/management/auth_files_provider_oauth.go  # helper + 5 handler call sites (item 8)

# 2. Build & tests
gofmt -l internal/api/ | wc -l                                # expect 0
go build -o /tmp/test-output ./cmd/server && rm /tmp/test-output
go test ./internal/api/... -count=1

# 3. Frontend locale structure — snippet from §3.5 (run inside frontend/)

# 4. Served bundle (repo root)
grep -c 'Management Center Info' static/management.html
grep -c 'do_not_use_proxy' static/management.html             # item 8 checkbox, expect >= 1
grep -c 'oauth-result-modal' static/management.html           # item 9 modal, expect >= 1

# 5. Compose deploys the fork build, not the upstream Docker Hub image
grep -n 'ghcr.io/arsydoni4326-alt/cliproxyapi' docker-compose.yml docker-compose.cluster.yml
```

If any check fails, reapply the corresponding item from §3 **before** committing
the merge.

---

## 5. Making future merges easy (recommended follow-ups)

1. **Commit the submodule locale fix** in
   `arsydoni4326-alt/Cli-Proxy-API-Management-Center` and bump the pointer in
   this repo. Uncommitted submodule changes are the most fragile state — any
   merge or submodule update discards them.
2. Consider upstreaming items 1–4 (they are generic robustness fixes, not
   fork-specific: graceful RESP error, TCP keep-alive, longer timeout). Once
   merged upstream, delete them from this document.
3. The GitHub token for the version check comes from `util.ResolveGitHubToken()`
   (env `GITHUB_TOKEN`/`GH_TOKEN`); if rate limiting (not timeouts) becomes the
   failure mode, set it rather than lengthening the timeout further.
4. Keep this document updated whenever items are added/removed, and reference it
   from the merge PR description so reviewers know the contract.

---

## 6. History

- **2026-09-12** — Initial version. Documents the four backend fixes and the
  frontend locale restructure introduced to fix cpa-usage-keeper connection
  resets, management-center raw i18n keys, and version-check timeouts
  (see `CHANGELOG.md` `[Unreleased]` and `session.md` of the same date).
- **2026-09-13** — Added item 7 (compose files must default to the fork image)
  and §7, a troubleshooting procedure for the `connection reset by peer`
  signature reported by cpa-usage-keeper. The compose defect was fixed (the fork
  fixes were present in the source tree but deployments pulled the upstream
  `eceasy/cli-proxy-api:latest` image by default). Later CPA-side logs showed the
  reset is **bidirectional and intermittent** — the CPA logged
  `redis protocol flush error: write tcp …: connection reset by peer` while the
  same `usage-queue` request returned `200` — so §7.1 and §7.4 were added to
  localize the reset on the connection path (kernel/firewall/NAT/proxy) instead
  of attributing it to the RESP handler. Parallel tcpdump on both hosts then
  proved **third-party TCP-RST injection**: each side sees an RST with the
  peer's address but **TTL 63** (both peers emit TTL 64), so neither program nor
  host generates them; §7.2 records the finding and the MAC-forensics step that
  names the injecting device.

---

## 7. Troubleshooting: cpa-usage-keeper reports `connection reset by peer`

Symptom, surfaced by the Keeper UI (ingest status `last_error`) and its logs:

```text
read redis subscribe auth response: read tcp <keeper>-><cpa>: read: connection reset by peer
usage: read redis queue auth response: read tcp <keeper>-><cpa>: read: connection reset by peer
request management usage queue: Get "http://<cpa>/v0/management/usage-queue?count=10000": read tcp <keeper>-><cpa>: read: connection reset by peer
```

Those three lines are the Keeper's startup probe (`internal/poller/redis_ingest_runner.go`
joins the subscribe, Redis-pull and HTTP-pull failures with `errors.Join`), so a single
line means one path failed while another worked, and all three together mean the CPA
side failed every entry point.

This is the client-side signature of a CPA build **without item 1 (§3.1)**: when
management routes are disabled, the old handler returned without writing a reply and
the kernel turned that close into a TCP RST.

### 7.1 Reading the CPA-side log

The same failure seen from the CPA looks like this:

```text
[error] [redis_queue_protocol.go:55] redis protocol flush error: write tcp <cpa>:8317-><keeper>: write: connection reset by peer
[info ] [gin_logger.go:103] 200 | 127ms | <keeper> | GET "/v0/management/usage-queue?count=10000"
```

`redis_queue_protocol.go:55` is the `flush()` closure of the command loop, i.e. the CPA
was *replying* (AUTH reply, SUBSCRIBE ack, LPOP payload or RESP error) and the far side
reset the socket underneath it. Note the line number does not identify the build: it is
line 55 before and after item 1 (the fix inserts at line ~68).

Interpretation:

- A flush error means the **client host/path** sent the RST first. The CPA handler cannot
  both write a reply and originate that reset.
- The follow-up `200` for the Keeper's own `usage-queue` request proves the CPA was up,
  `managementRoutesEnabled` was true, and the Keeper's management key authenticated. The
  very request the Keeper reported as reset succeeded one second later, so failures here
  are intermittent rather than caused by a missing route or a rejected key.
- When the CPA resets *and* the Keeper resets on fresh handshakes, the reset is being
  produced on the connection path (kernel/firewall/NAT/proxy), not by either program.

Checks, in order:

1. **What is the CPA actually running?** On the CPA host:

   ```bash
   docker inspect --format '{{.Config.Image}}' cli-proxy-api
   docker logs cli-proxy-api 2>&1 | head -1     # version/commit banner
   ```

   `eceasy/cli-proxy-api:latest` is the upstream image and does **not** contain fork
   fixes. The fork image `ghcr.io/arsydoni4326-alt/cliproxyapi:latest` is built from
   this repository's `Dockerfile` on every push to `main` (tag `sha-da05d5e` and
   `latest` currently point at the same digest), and contains items 1–6.
2. **Probe the RESP side directly** from the Keeper host:

   ```bash
   printf '*1\r\n$4\r\nPING\r\n' | timeout 3 nc <cpa-host> 8317
   ```

   - `-ERR remote management disabled` → the fix is present, but management is not
     available to this client: set `remote-management.secret-key` and, for a Keeper on
     another host, `remote-management.allow-remote: true`, then restart CPA and make
     sure `CPA_MANAGEMENT_KEY` in the Keeper `.env` holds that same secret.
   - `-NOAUTH Authentication required.` → management is enabled; the handshake works,
     so a reset here is not produced by the RESP handler (see step 3).
   - No output, or `nc: connection reset by peer` → **the running build predates the
     fix**; redeploy the fork image (§3.7) or rebuild `./Dockerfile` and restart.
3. **If the probe is clean but resets keep coming**, the failure originates outside the
   RESP handler. The usual cause is the CPA process or container restarting: every Keeper
   entry point fails in the same burst, and connections that race the shutdown surface as
   a reset or as an unexpected EOF. Confirm with
   `docker ps --format '{{.Names}} {{.Status}}'` (uptime / restart count) and
   `docker logs --since 15m cli-proxy-api` (crash, OOM kill, or container recreate).
4. **Confirm where the RST is produced** when both sides log a reset. Run the raw probe
   repeatedly from the Keeper host, then capture traffic on both hosts at the same time:

   ```bash
   # Keeper host: raw RESP probe, 20 attempts
   for i in $(seq 1 20); do
     printf '*1\r\n$4\r\nPING\r\n' | timeout 2 nc 192.168.1.12 8317 | head -c 60
     echo "  <- attempt $i"
     sleep 0.5
   done

   # CPA host and Keeper host, in parallel:
   sudo tcpdump -nni any host 192.168.1.10 and port 8317
   ```

   Read the capture for who emits `R` (RST) and on which side of the flow. On the Keeper
   host also check the path, because a proxy/VPN or subnet router between the two hosts
   produces exactly this signature:

   ```bash
   ip route get 192.168.1.12
   ip rule; sudo nft list ruleset 2>/dev/null | grep -iE 'tproxy|mark'
   ```

   On the CPA host check that a firewall or conntrack table is not rejecting bursts
   (iptables `REJECT --reject-with tcp-reset` and a full conntrack table both surface as
   `connection reset by peer` at the client):

   ```bash
   sudo iptables-save | grep -n 8317
   sudo conntrack -L 2>/dev/null | grep -c 8317
   sudo dmesg -T | grep -iE 'conntrack|nf_conntrack' | tail
   ```

   The Keeper opens a fresh Redis TCP connection for every pull, so a rate-limited
   firewall or an undersized conntrack table is hit by its reconnect bursts rather than
   by steady traffic — which is why a single `usage-queue` request can succeed while the
   three handshake paths fail.

### 7.2 What the packet captures prove (2026-09-13)

Captures taken on both hosts (`tcpdump -nni any -vv 'host <peer> and port 8317'`) show
the exchange working and then being killed from outside:

```text
# CPA host (192.168.1.12):
.10.52756 > .12.8317: Flags [S]   ttl 64   <- Keeper's own stack
.12.8317 > .10.52756: Flags [S.]  ttl 64   <- CPA's own stack
... full handshake + PUSH/ACK both ways ...
.12.8317 > .10.52756: Flags [P.]  ttl 64   <- CPA AUTH reply (or LPOP/usage-queue JSON)
.10.52756 > .12.8317: Flags [R]   ttl 63   <- "from the Keeper", but TTL 63!
# Keeper host (192.168.1.10), same exchange, mirrored:
.12.8317 > .10.52756: Flags [R]   ttl 63   <- "from the CPA", but TTL 63!
```

Two facts identify this as *third-party TCP-RST injection* on the LAN:

- **Spoofed source.** Each host sees an RST attributed to the peer, but the **peer's own
  packets carry TTL 64** (both are Linux, direct WiFi, single L2 hop). An RST with
  **TTL 63** cannot have been produced by the peer's stack — it passed through one extra
  L3 hop, i.e. it was fabricated by a device that impersonates the peer (typically the
  default gateway / WiFi router, `192.168.1.1`, which is one hop from every WLAN client).
- **Bidirectional, timed to the reply.** The fake `.10 → .12` RST arrives ~6 ms after the
  CPA sends each response (AUTH reply, 714-byte `usage-queue` JSON); the fake `.12 → .10`
  RST follows within a few hundred ms. The Keeper then logs `read …: connection reset by
  peer` and the CPA logs `redis protocol flush error: write …: connection reset by peer`.

The 200 response did arrive at `.10` in the same capture window before the RST —
everything below the RST injection is healthy. Neither application nor its configuration
is at fault.

Typical sources of exactly this signature: a router/AP running an intrusion-prevention /
"DoS protection" / QoS-assist feature, or a LAN box running a transparent proxy /
Tailscale "exit node"/subnet-router or TPROXY-style policy routing that resets flows it
does not match.

**Identify the injector by MAC.** Re-run the Keeper probe while capturing with link
headers, and compare the RST frames' source MAC with the router's MAC:

```bash
# on the CPA host:
sudo tcpdump -nnei wlp3s0 -c 20 'tcp[13] & 4 != 0 and host 192.168.1.10'
# on the Keeper host:
sudo tcpdump -nnei wlp2s0 -c 20 'tcp[13] & 4 != 0 and host 192.168.1.12'
```

The RST frames' source MAC is the injector. Then disable the corresponding feature
(router IPS/DoS-protection/AiProtection/QoS-assist, or the transparent-proxy/
Tailscale-exit-node policy rules on that host), add a `direct` rule for `192.168.1.0/24`,
and re-test. A temporary wired (dumb-switch) path between the two hosts proves the
diagnosis if the resets disappear.
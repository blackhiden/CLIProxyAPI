# Session: Test Credential Models Feature

**Date:** 2026-11-17
**Status:** Merged into local `main`; feature worktrees retained

## Objective and work split

Implement per-model auth-file credential verification through the existing v8
credential-pinned verifier. Frontend work is on
`feature/test-credential-models-frontend`; documentation, merge preservation,
and the generated root management asset are on
`feature/test-credential-models-docs`. Both root branches originate from
`main` commit `3b1c7a688222fefbd51c6099f35ffed1b7d34fd0`. Both are integrated
into local `main`; no remote push was performed.

## Decisions and discoveries

- The backend already exposes `POST /v8/management/credentials/verify` and
  applies `SanitizeRenewalMessage`; no backend implementation change was needed.
- The frontend sends selected credential filename and exact model ID. Results
  are modal-local, per-model state and clear when credentials switch or the
  modal closes. The button stops row click propagation to preserve copy.
- Verification performs a real minimal inference and can consume upstream
  quota; the UI makes no claim that this is a passive or zero-cost probe.
- Local roadmap planning files are ignored by Git and absent from `main`;
  they were left untouched. The feature has been documented in tracked
  `docs/MERGE-PRESERVATION.md` and this session file.
- The required root `SCHEMA.md` was absent and is now added.

## Validation

- `bun test tests/managementV8Api.test.ts`: 14 pass.
- `bun run type-check`: pass.
- `bun run lint`: exit 0; four existing warnings in unrelated tests.
- `bun run build`: pass; the built single-file artifact includes Test Model UI.
- `bun run verify`: the focused feature tests, lint, type-check, and build pass;
  the full suite has 2388 pass, 1 todo, and 18 pre-existing failures in locale
  completeness, LoginRevealVeil, and visual-config rebase suites.
- Backend targeted tests pass: `internal/api/handlers/management` and
  `sdk/cliproxy/auth`. The required server build also passes.

## Integration result

- Root asset/preservation checks pass. The docs branch has an uncommitted,
  pre-existing `cpauk` submodule checkout mismatch from initialization; it was
  not staged or changed.
- The frontend and documentation feature branches were merged in a temporary
  integration worktree from shared base `3b1c7a688222fefbd51c6099f35ffed1b7d34fd0`.
  Local `main` was fast-forwarded to `961b94af25429c0e39ad33cb87aebd5ebad314ca`.
- A final continuity update was committed on local `main` as
  `f8eee9dd31f3d104d1b8eed308c5ab06aa3daa8d`.
- Integration reruns passed: frontend API contract tests (14), frontend type
  check and build, root server build, and generated-asset/preservation checks.
- No remote push was performed. Feature branches and their worktrees remain
  intact. The uncommitted `cpauk` worktree mismatch is preserved and excluded
  from all commits.
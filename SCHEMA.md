# Data and API Schema Overview

CLIProxyAPI stores provider credentials and configuration through the storage
abstractions documented in `AGENTS.md` and `ARCHITECTURE.md`. The default auth
material directory is `auths/`; optional backends include Postgres, Git, and
object storage.

## Management API

- The management API exposes v8 endpoints beneath `/v8/management`.
- Auth credential files are listed through `/credentials` and their supported
  models through `/credentials/models`.
- Credential model verification is a non-persistent operator action using
  `POST /credentials/verify` with a credential `name` and exact `model`
  identifier. It executes a minimal, credential-pinned inference through
  `Manager.VerifyInference`; it is not a passive metadata check and may consume
  upstream quota.
- Verification results are transient frontend state and are not stored in a database.

## Persistence

No database schema or migration is introduced by per-model credential testing. The feature stores no verification result beyond the lifetime of the open frontend models modal.
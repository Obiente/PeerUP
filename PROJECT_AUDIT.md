# Uppe Project Audit

Date: 2026-04-04

## Executive Summary

Uppe is a decentralized uptime monitoring platform built around a Rust service that performs checks, stores results locally, and participates in a libp2p-based peer network. A Go API server exposes monitoring data over ConnectRPC for a frontend dashboard built in Astro/TypeScript.

The project direction is coherent, but the implementation is in a transitional state:

- The Rust service is the most complete part and compiles successfully.
- The Go API server compiles and `go test ./...` passes, but it only implements monitor/result services.
- The frontend has real integration for core monitor/result views, but several pages and services are still placeholders or mock-backed.
- The codebase has architectural drift around schema ownership, API surface area, and frontend messaging about P2P readiness.

## Verified State

- `cargo check -p uppe-service`: passes
- `go test ./...` in `apps/server`: passes
- Frontend runtime/tests were not fully verified in this shell because Node is not available on the Linux path even though `pnpm` resolves.

## What The Project Is

Uppe is intended to be a peer-to-peer uptime monitoring system:

- Users create monitors for HTTP, TCP, and eventually ICMP targets.
- The Rust service executes checks on schedule and stores results in a shared SQLite/LibSQL database.
- Results can be signed and shared across peers through the PeerUP/libp2p layer.
- The Go server reads from the shared database and exposes API endpoints for dashboards and other clients.
- The Astro frontend renders dashboards, analytics, service detail pages, network views, settings, and public status pages.

## How It Works Today

### 1. Core runtime

- The main runtime entrypoint is [`apps/service/src/main.rs`](/home/crunchy/dev/Obiente/Uppe/apps/service/src/main.rs).
- The service loads config, initializes the shared LibSQL database, loads a keypair, sets up location tracking, and starts the orchestrator.
- The orchestrator in [`apps/service/src/orchestrator/mod.rs`](/home/crunchy/dev/Obiente/Uppe/apps/service/src/orchestrator/mod.rs) coordinates monitoring, database writes, P2P startup, helper assignment logic, and retention cleanup.

### 2. Monitoring flow

- Monitor definitions live in the `monitors` table.
- Scheduled checks produce `monitor_results`.
- Peer-received results are stored separately in `peer_results`.
- The monitoring engine currently supports HTTP/HTTPS and TCP. ICMP is still a stub in [`apps/service/src/monitoring/checker.rs#L88`](/home/crunchy/dev/Obiente/Uppe/apps/service/src/monitoring/checker.rs#L88).

### 3. API layer

- The Go server entrypoint is [`apps/server/cmd/server/main.go`](/home/crunchy/dev/Obiente/Uppe/apps/server/cmd/server/main.go).
- The server registers only `MonitorService` and `ResultService` in [`apps/server/internal/server/server.go#L66`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/server/server.go#L66).
- The monitor and result services read from the shared database and expose ConnectRPC endpoints used by the client dashboard.

### 4. Frontend

- The frontend uses generated Connect clients in [`apps/client/src/lib/api/client.ts`](/home/crunchy/dev/Obiente/Uppe/apps/client/src/lib/api/client.ts).
- Dashboard and analytics pages pull real monitor/result data from the Go API.
- Several other areas are still static or mock-backed: settings, network map, public status pages, and status-page management.

## High-Priority Findings

### 1. Competing database schemas and multiple "sources of truth"

Severity: High

The repo currently has at least three schema definitions that do not agree:

- Rust migration source in [`apps/service/src/database/migrations.rs#L7`](/home/crunchy/dev/Obiente/Uppe/apps/service/src/database/migrations.rs#L7)
- Shared schema doc in [`shared/database/schema.sql#L1`](/home/crunchy/dev/Obiente/Uppe/shared/database/schema.sql#L1)
- Old Go migration file in [`apps/server/internal/db/migrations/001_initial_schema.sql#L1`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/db/migrations/001_initial_schema.sql#L1)

Examples of drift:

- Shared schema says `monitors.id` is `INTEGER PRIMARY KEY AUTOINCREMENT` in [`shared/database/schema.sql#L24`](/home/crunchy/dev/Obiente/Uppe/shared/database/schema.sql#L24).
- Old Go migration says `monitors.id` is `TEXT PRIMARY KEY` and uses `url`/`type` columns in [`apps/server/internal/db/migrations/001_initial_schema.sql#L6`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/db/migrations/001_initial_schema.sql#L6).
- Rust migrations actually create `target`/`check_type` and append newer orchestration fields in [`apps/service/src/database/migrations.rs#L97`](/home/crunchy/dev/Obiente/Uppe/apps/service/src/database/migrations.rs#L97).

Impact:

- This is the biggest architecture risk in the repo.
- New contributors can follow the wrong schema.
- Any future non-Rust migration work can silently break compatibility.

Recommendation:

- Make Rust migrations the only executable schema source.
- Mark the old Go migration directory as obsolete or remove it.
- Generate `shared/database/schema.sql` from the Rust schema, or delete it if it cannot be kept authoritative.

### 2. The Go server exposes only part of the declared API surface

Severity: High

The proto layer defines `NetworkService`, `SettingsService`, and `StatusPageService`:

- [`apps/server/proto/network/v1/network.proto#L9`](/home/crunchy/dev/Obiente/Uppe/apps/server/proto/network/v1/network.proto#L9)
- [`apps/server/proto/settings/v1/settings.proto#L7`](/home/crunchy/dev/Obiente/Uppe/apps/server/proto/settings/v1/settings.proto#L7)
- [`apps/server/proto/statuspage/v1/statuspage.proto#L9`](/home/crunchy/dev/Obiente/Uppe/apps/server/proto/statuspage/v1/statuspage.proto#L9)

But the server registers only monitor/result handlers in [`apps/server/internal/server/server.go#L66-L75`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/server/server.go#L66-L75), and there are no implementations for those additional services.

Impact:

- The API contract implies more capability than the backend actually provides.
- Frontend and docs can drift ahead of reality.

Recommendation:

- Either implement and register these services next, or remove/defer the protos until the backend contract is real.

### 3. Frontend product claims do not match backend reality

Severity: High

Core examples:

- Network map page says P2P is "coming soon" and uses generated fake node data in [`apps/client/src/pages/network-map.astro#L25-L45`](/home/crunchy/dev/Obiente/Uppe/apps/client/src/pages/network-map.astro#L25-L45), even though the Rust service already contains substantial P2P code.
- Settings page is entirely local defaults and not wired to any settings backend in [`apps/client/src/pages/settings.astro#L24-L53`](/home/crunchy/dev/Obiente/Uppe/apps/client/src/pages/settings.astro#L24-L53).
- Public status page route is static/mock content in [`apps/client/src/pages/public/[slug].astro#L18-L77`](/home/crunchy/dev/Obiente/Uppe/apps/client/src/pages/public/[slug].astro#L18-L77).
- Status page service is still returning fabricated data in [`apps/client/src/lib/services/statusPageService.ts#L67-L287`](/home/crunchy/dev/Obiente/Uppe/apps/client/src/lib/services/statusPageService.ts#L67-L287).

Impact:

- The system message to users is inconsistent.
- The frontend currently overstates readiness in some docs and understates it in some pages.

Recommendation:

- Decide the product truth for this phase:
  - Either "P2P core exists, UI is incomplete"
  - Or "P2P is experimental, not user-ready"
- Then align pages, docs, and API surfacing to that one statement.

### 4. Signature verification is still bypassed in critical trust paths

Severity: High

Examples:

- PeerUP distributed data verification returns `true` unconditionally in [`crates/peerup/src/distributed/messages.rs#L100-L105`](/home/crunchy/dev/Obiente/Uppe/crates/peerup/src/distributed/messages.rs#L100-L105).
- Admin trust-chain verification uses `signature.len() == 64` as a placeholder in [`apps/service/src/orchestrator/admin_trust.rs#L113-L167`](/home/crunchy/dev/Obiente/Uppe/apps/service/src/orchestrator/admin_trust.rs#L113-L167).

Impact:

- The decentralized trust model is not yet enforceable.
- Architectural claims around signed results and key-chain validation are only partially true.

Recommendation:

- Treat this as a security milestone, not a cleanup task.
- Implement real canonical-message signing and verification before expanding network-facing features.

## Medium-Priority Findings

### 5. `NewDatabase` can panic on short database paths

Severity: Medium

In [`apps/server/internal/db/factory.go#L21`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/db/factory.go#L21), the code slices `databaseURL[:7]` and `databaseURL[:8]` without guarding length first.

Impact:

- A short path like `a.db` can panic before the server returns a configuration error.

Recommendation:

- Replace manual slicing with `strings.HasPrefix`.

### 6. CORS handling is unsafe and invalid for credentialed requests

Severity: Medium

[`apps/server/internal/server/server.go#L27-L39`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/server/server.go#L27-L39) reflects arbitrary origins and also sets `Access-Control-Allow-Credentials: true`. It also falls back to `*`.

Impact:

- `*` with credentials is invalid.
- Reflecting arbitrary origins is not a real production CORS policy.

Recommendation:

- Add an allowlist from config.
- Disable credentials unless there is a specific authenticated browser flow requiring them.

### 7. Go model layer is duplicated and internally inconsistent

Severity: Medium

There are at least three overlapping model representations:

- Business model in [`apps/server/internal/models/monitor.go`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/models/monitor.go)
- "Shared schema" DB model in [`apps/server/internal/models/shared_schema.go`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/models/shared_schema.go)
- Older GORM model in [`apps/server/internal/models/monitor_gorm.go`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/models/monitor_gorm.go)

These disagree on:

- `id` type
- whether `uuid` and `id` are the same
- whether the DB uses `target` vs `url`
- whether visibility fields are included

Impact:

- This raises maintenance cost and schema confusion.
- The current code compiles, but future changes are likely to update the wrong model.

Recommendation:

- Collapse onto one DB model layer and one business model layer.
- Remove `monitor_gorm.go` if `shared_schema.go` is the real path forward.

### 8. Settings page hardcodes the wrong API port

Severity: Medium

[`apps/client/src/pages/settings.astro#L86-L89`](/home/crunchy/dev/Obiente/Uppe/apps/client/src/pages/settings.astro#L86-L89) displays `http://localhost:8081`, while the actual default API base is `http://localhost:8080` in [`apps/client/src/lib/api/client.ts#L12`](/home/crunchy/dev/Obiente/Uppe/apps/client/src/lib/api/client.ts#L12) and the Go server default is port 8080 in [`apps/server/internal/config/config.go`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/config/config.go).

Impact:

- This is a user-visible correctness bug.

Recommendation:

- Drive displayed endpoint information from the same config/env source the client uses.

### 9. ICMP is exposed as a monitor type but is not implemented

Severity: Medium

The system still exposes ICMP as a real monitor type in multiple places, but the checker explicitly returns an error in [`apps/service/src/monitoring/checker.rs#L88-L109`](/home/crunchy/dev/Obiente/Uppe/apps/service/src/monitoring/checker.rs#L88-L109).

Impact:

- Users can think ping monitoring exists when it does not.

Recommendation:

- Either remove ICMP from user-facing creation flows for now or mark it as experimental/disabled in UI and API.

## Lower-Priority / Structural Findings

### 10. Repo hygiene needs tightening

Observed artifacts include:

- Local databases under `shared/data/`
- Multiple keypair files under `apps/service/*.key`
- Generated Astro workspace artifacts in `apps/client/.astro/`

Some are ignored, but keypair files are currently present in the worktree and should not live as normal project assets.

Recommendation:

- Ignore generated keypair files and local dev databases consistently.
- Keep sample fixtures distinct from real runtime credentials.

### 11. Frontend docs are inconsistent

- [`apps/client/README.md`](/home/crunchy/dev/Obiente/Uppe/apps/client/README.md) is still the Astro starter README.
- Other client docs claim production readiness or real-data integration.

Recommendation:

- Replace the starter README with an accurate frontend overview.

### 12. There is little automated test coverage around the riskiest logic

What was verified:

- Go packages compile and tests pass, but there are effectively no substantive tests there.
- Rust service compiles.

What is still weak:

- No meaningful audit-proof tests for schema compatibility.
- No strong end-to-end tests for Rust service + Go API + shared DB.
- No security tests around signature verification or trust-chain logic.

## Unimplemented Features

### Backend/API

- `NetworkService`: proto exists, server implementation absent.
- `SettingsService`: proto exists, server implementation absent.
- `StatusPageService`: proto exists, server implementation absent.
- `ResultService.StreamResults`: explicit TODO in [`apps/server/internal/connect/result/result.go#L100-L113`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/connect/result/result.go#L100-L113).
- PostgreSQL/TimescaleDB support: explicitly unimplemented in [`apps/server/internal/db/factory.go#L27-L29`](/home/crunchy/dev/Obiente/Uppe/apps/server/internal/db/factory.go#L27-L29).

### Rust service

- Real ICMP monitoring.
- Real admin-trust signature verification.
- Full DHT-backed admin key caching/rotation completion.
- Some retention/orchestration paths remain placeholders.

### Frontend

- Real settings integration.
- Real network map integration.
- Real public status pages.
- Real status-page CRUD and publish flow.
- Incidents and alerting are still not backed end-to-end.

## Structure Assessment

## What is structured well

- The top-level split is sensible:
  - `apps/service` for the runtime node
  - `apps/server` for API
  - `apps/client` for UI
  - `crates/peerup` for reusable P2P logic
- The service/orchestrator boundary is reasonable.
- The Go API reading a shared DB instead of controlling the monitoring engine is the right separation.

## What is not structured well enough yet

- Schema authority is not singular in practice.
- The Go model layer is duplicated.
- Product surface is defined in more places than implementation exists.
- The frontend includes multiple "future" feature surfaces as if they were near-complete, while the backend contracts for those features are still mostly declarative.

## Recommended Next Steps

### Phase 1: Stabilize the architecture

1. Make Rust migrations the only schema source of truth.
2. Remove or clearly archive obsolete Go migration/schema files.
3. Consolidate Go DB/business models into one path.
4. Replace the Astro starter README and align all docs with actual feature status.

### Phase 2: Close correctness/security gaps

1. Fix the database path prefix panic in `apps/server/internal/db/factory.go`.
2. Replace permissive CORS reflection with configured origins.
3. Implement real signature verification for:
   - peer data
   - admin key rotations
   - revocation lists
4. Hide or disable ICMP until it actually works.

### Phase 3: Finish the missing API surface

1. Implement `SettingsService`.
2. Implement `NetworkService`.
3. Implement `StatusPageService`.
4. Decide whether `StreamResults` is actually required now; if not, remove it from the short-term contract.

### Phase 4: Bring the frontend back in sync

1. Replace placeholder settings page with live API-backed values.
2. Replace the network map placeholder with real peer/network stats or remove the page from nav until it exists.
3. Replace mock public status pages with backend-backed content.
4. Remove any remaining mock utility modules that are no longer part of the intended runtime path.

### Phase 5: Add integration confidence

1. Add an integration test that boots:
   - Rust service migrations
   - Go API
   - shared DB assertions
2. Add contract tests for schema compatibility.
3. Add end-to-end frontend smoke tests for:
   - dashboard
   - analytics
   - monitor CRUD

## Bottom Line

The project is not off-plan conceptually. The intended architecture is still sound:

- Rust node as the source of monitoring and P2P truth
- Go API as a read/write presentation layer
- Astro frontend as a dashboard and public UI

Where it is drifting is execution discipline:

- too many parallel "source of truth" files
- too many declared APIs without implementations
- too many placeholder UIs presented alongside real features

If the next work focuses on consolidation before expansion, the architecture can still hold cleanly. If new features keep landing before the schema/API/documentation layers are unified, the project will become much harder to evolve safely.

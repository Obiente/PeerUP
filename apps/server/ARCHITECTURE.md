# Uppe. Architecture - Decentralized Monitoring Network

## Core Principle

**Uppe. is a decentralized monitoring network where every peer strengthens the network.**

- **All instances JOIN the decentralized network by default** - mobile phones, self-hosters, servers
- **Self-hosters join the network** - they participate in and strengthen the decentralized network
- **Rust Service (PeerUP/P2P) is self-contained** - each peer operates autonomously
- **P2P Network is the core** - peers share monitoring results, strengthening the network
- **Optional isolation** - can run without P2P (loses decentralized benefits, but still works locally)
- **Go API is optional** - only needed when you want a local web frontend for data visibility
- **No central authority** - each peer is equal and autonomous

## Design Principles

1. **Decentralized**: Every peer is autonomous and contributes to network strength
2. **P2P-First**: Peer-to-peer communication is the core, not centralized APIs
3. **Self-Contained Core**: Rust service (monitoring + P2P) works independently
4. **Network Resilience**: More peers = stronger network (no single point of failure)
5. **Optional Frontend API**: Go API is only for local frontend data visibility
6. **Mobile-First**: Must run efficiently on mobile devices (battery, CPU, memory)
7. **Zero-Config**: Easy deployment for thousands of peers
8. **Scale Adaptive**: Works from single device to large clusters
9. **Performance**: Minimal overhead, direct data access
10. **Flexibility**: Support multiple deployment modes

## Deployment Modes

**All modes can join the decentralized P2P network - every peer strengthens the network, including mobile phones and self-hosters.**

**Default: Join the network** - All instances participate in the decentralized monitoring network by default.
**Optional: Isolated mode** - Can run without P2P (loses decentralized benefits).

The only difference between modes is whether you want a local web frontend (Go API).

### Mode 1: Rust Service Only (Decentralized Peer)

**Self-contained Rust service - Perfect for mobile phones, full P2P network participation**

```
┌─────────────────────────────────────┐
│     Rust Service (Decentralized Peer)│
│  ┌──────────────────────────────┐  │
│  │  Monitoring Engine            │  │
│  │  - Reads monitors from DB     │  │
│  │  - Performs checks            │  │
│  │  - Writes results to DB       │  │
│  └──────────────┬───────────────┘  │
│                 │                   │
│  ┌──────────────▼───────────────┐  │
│  │  PeerUP P2P Network          │  │
│  │  - Kademlia DHT              │  │
│  │  - mDNS Discovery             │  │
│  │  - Peer Communication          │  │
│  │  - Shares results with peers   │  │
│  │  - Receives results from peers │  │
│  │  - Strengthens network         │  │
│  └──────────────┬───────────────┘  │
│                 │                   │
│  ┌──────────────▼───────────────┐  │
│  │  Embedded SQLite (libsql)    │  │
│  │  - Zero external deps        │  │
│  │  - Single file database       │  │
│  │  - Direct read/write access   │  │
│  │  - Stores local + peer data   │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
         │
         │ P2P Network
         │
    ┌────┴────┐
    │  Other  │
    │  Peers  │
    └─────────┘
```

**Characteristics:**
- ✅ **Fully self-contained** - no external dependencies
- ✅ **Full P2P network peer** - equal participant, strengthens network
- ✅ **Mobile-friendly** - optimized for mobile phones (battery, CPU, memory)
- ✅ **Shares monitoring results** - contributes to network resilience
- ✅ **Receives peer results** - benefits from redundant monitoring
- ✅ **Direct database access** - reads/writes directly
- ✅ **Embedded SQLite** - zero external database needed
- ✅ **Minimal resource usage** (~50-100 MB memory, perfect for mobile)
- ✅ **Works offline** - can operate without internet (local monitoring)
- ✅ **P2P networking** - full participation in decentralized network
- ✅ **No API needed** - operates independently
- ✅ **Network resilience** - more peers = stronger network

**Network Participation:**
- ✅ **Full peer** - mobile phones are equal peers in the network
- ✅ **Monitors URLs** - assigned monitoring tasks
- ✅ **Shares results** - cryptographically signed results shared peer-to-peer
- ✅ **Receives results** - gets results from other peers (redundant monitoring)
- ✅ **Strengthens network** - every peer makes the network more resilient
- ✅ **No limitations** - mobile phones have full P2P capabilities

**Use Cases:**
- Mobile phones (primary use case - full network participation)
- Raspberry Pi
- Edge devices
- Personal monitoring nodes
- P2P network participants
- Any device that wants to contribute to the network

**Configuration:**
```toml
[mode]
type = "standalone"  # Self-contained, no API

[database]
type = "libsql"
path = "./uppe.db"

[peerup]
enabled = true
# No API configuration needed
```

### Mode 2: Rust Service + Go API (Self-Hosted with Frontend)

**Self-hosters JOIN the decentralized network (Rust) + Optional local frontend (Go API)**

**Default: Joins the decentralized P2P network** - Self-hosters participate in the large monitoring network by default, strengthening it with their contribution.

**Optional: Isolated mode** - Can run without P2P network (loses decentralized benefits, but still works locally).

**Note: Rust service is a full peer in the P2P network by default, just like Mode 1. Go API is only for local frontend visibility.**

```
┌─────────────────┐         ┌─────────────────────────────┐
│   Go API        │         │  Rust Service               │
│   (apps/server) │         │ (Decentralized Peer)         │
│                 │         │                             │
│  - Local API    │         │  - Monitoring (self-contained)│
│  - Data queries │         │  - P2P Network               │
│  - Dashboards   │         │  - Shares results with peers │
│  - Auth         │         │  - Direct DB writes          │
│  (Local only)   │         │  - Strengthens network       │
└────────┬────────┘         └────────┬────────────────────┘
         │                           │
         │  Reads from shared DB     │  Writes to shared DB
         │                           │
         └───────────┬───────────────┘
                     │
         ┌───────────▼───────────┐
         │   Shared Database      │
         │   - SQLite (small)     │
         │   - PostgreSQL (large) │
         │                         │
         │  Rust: Read/Write       │
         │  Go API: Read-only      │
         └────────────────────────┘
         │
         │ P2P Network
         │
    ┌────┴────┐
    │  Other  │
    │  Peers  │
    └─────────┘
```

**Characteristics:**
- ✅ **Rust service is full P2P network peer** - operates independently, strengthens network
- ✅ **Same P2P capabilities as Mode 1** - full network participation
- ✅ **Go API is optional** - only for local frontend data visibility
- ✅ **Shared database** - both read from same DB
- ✅ **No coupling** - Rust service doesn't call Go API
- ✅ **Direct database access** - no network overhead
- ✅ **Full P2P network participation** - shares results with other peers
- ✅ **Separation of concerns** - monitoring vs frontend

**Data Flow:**
- **Rust Service** (Full P2P Peer): 
  - Reads monitors → Performs checks → Writes results (all direct DB)
  - Shares results with peers via P2P network
  - Receives results from peers via P2P network
  - **Same network participation as Mode 1**
- **Go API** (Local Frontend Only): 
  - Reads monitors/results → Serves to local frontend (read-only DB access)
  - Does not affect P2P network participation

**Network Benefits:**
- ✅ **Full peer** - participates in decentralized network (same as Mode 1)
- ✅ **Shares results** - monitoring results shared with other peers
- ✅ **Receives results** - gets results from other peers (redundant monitoring)
- ✅ **Strengthens network** - contributes to network resilience
- ✅ **Local visibility** - Go API provides local frontend (doesn't affect network)

**Network Participation:**
- ✅ **Joins decentralized network by default** - participates in large monitoring network
- ✅ **Strengthens network** - contributes monitoring results to network
- ✅ **Receives network benefits** - gets results from other peers (redundant monitoring)
- ✅ **Full peer** - equal participant in decentralized network
- ✅ **Optional isolation** - can disable P2P to run isolated (loses network benefits)

**Use Cases:**
- Self-hosted deployments with web UI (joins network by default)
- Users who want dashboards and analytics
- Medium-scale monitoring
- Peers who want local visibility into their contribution
- Users who want to strengthen the decentralized network

**Configuration:**
```toml
# Go API (optional, only for frontend)
[mode]
type = "api-server"  # Frontend API only

[database]
type = "libsql"  # or "postgresql"
path = "./uppe.db"  # Read-only access recommended

# Rust Service (self-contained)
[mode]
type = "standalone"  # Self-contained, no API dependency

[database]
type = "libsql"
path = "./uppe.db"  # Same path, read/write access
```

### Mode 3: Distributed Decentralized Network (Large Scale)

**The decentralized network itself - all peers (mobile phones, self-hosters, servers) join this network**

**This is the network that Mode 1 and Mode 2 peers join** - it's not a separate mode, it's the network itself.

**Note: Every instance (Mode 1, Mode 2) joins this decentralized network by default. The network is made up of all participating peers.**

```
┌─────────────────┐
│  Rust Service   │───P2P Network───┐
│  (Peer 1)       │                  │
│  - Decentralized│                  │
│  - Direct DB    │                  │
│  - Strengthens   │                  │
│    network      │                  │
└────────┬────────┘                  │
         │                            │
         │  Direct DB Access          │
         │                            │
┌────────▼────────┐                  │
│  Local DB (Peer 1)│                  │
└─────────────────┘                  │
                                     │
┌─────────────────┐                  │
│  Rust Service   │───P2P Network───┤
│  (Peer 2)       │                  │
│  - Decentralized│                  │
│  - Direct DB    │                  │
│  - Strengthens   │                  │
│    network      │                  │
└────────┬────────┘                  │
         │                            │
         │  Direct DB Access          │
         │                            │
┌────────▼────────┐                  │
│  Local DB (Peer 2)│                  │
└─────────────────┘                  │
                                     │
┌─────────────────┐                  │
│  Rust Service   │───P2P Network───┤
│  (Peer N)       │                  │
│  - Decentralized│                  │
│  - Direct DB    │                  │
│  - Strengthens   │                  │
│    network      │                  │
└────────┬────────┘                  │
         │                            │
         │  Direct DB Access          │
         │                            │
┌────────▼────────┐                  │
│  Local DB (Peer N)│                  │
└─────────────────┘                  │
                                     │
                    ┌─────────────────┴──────────────┐
                    │   Go API Server (Optional)     │
                    │   (apps/server)                │
                    │  - Aggregates from P2P network │
                    │  - Serves frontend              │
                    │  - Auth & Rate limiting         │
                    └───────────┬────────────────────┘
                                │
                    ┌───────────▼───────────┐
                    │   Aggregation DB      │
                    │   PostgreSQL/TimescaleDB│
                    │   - Aggregated data     │
                    │   - Time-series optimized│
                    │   (Optional, for UI)    │
                    └────────────────────────┘
```

**Characteristics:**
- ✅ **Each Rust service is a full P2P network peer** - operates independently
- ✅ **Same P2P capabilities as Mode 1 & 2** - full network participation per peer
- ✅ **P2P network is the core** - peers communicate peer-to-peer
- ✅ **Local databases** - each peer has its own DB
- ✅ **Network strength** - more peers = stronger, more resilient network
- ✅ **No central authority** - all peers are equal (mobile phones = servers = equal)
- ✅ **Go API is optional** - aggregates data for frontend (not required for network)
- ✅ **Horizontally scalable** - add more peers (mobile phones, servers, etc.) to strengthen network
- ✅ **Best for high throughput** - distributed monitoring
- ✅ **All peers equal** - mobile phones participate same as servers

**Data Flow:**
- **Rust Services (Full P2P Peers)**: 
  - Each operates independently, writes to local DB
  - Monitors assigned URLs
  - Shares results with peers via P2P network
  - Receives results from peers via P2P network
  - Strengthens network with each contribution
  - **Same network participation regardless of device** (mobile phone = server = equal peer)
- **P2P Network**: 
  - Decentralized communication between peers
  - All peers are equal (mobile phones, servers, edge devices)
  - Cryptographically signed results
  - Redundant monitoring (multiple peers monitor same URLs)
  - No single point of failure
  - Network strength grows with each peer (including mobile phones)
- **Go API (Optional)**: 
  - Aggregates data from P2P network for frontend
  - Not required for network operation
  - Provides centralized view of decentralized data
  - Does not affect peer participation

**Network Benefits:**
- **Decentralized**: No single point of failure
- **Resilient**: Network strength grows with each peer (including mobile phones)
- **Redundant**: Multiple peers monitor same URLs
- **Trust**: Cryptographically signed results
- **Scalable**: Add peers (mobile phones, servers, edge devices) to strengthen network
- **Equal Participation**: All devices are equal peers (mobile phones = servers)

**Use Cases:**
- Large-scale decentralized monitoring network
- Hosted Uppe. product (aggregation layer)
- Multi-region deployments
- Enterprise customers with many peers
- Community-driven monitoring networks

**Configuration:**
```toml
# Go API (frontend aggregator)
[mode]
type = "api-server"

[database]
type = "postgresql"
url = "postgresql://..."  # Centralized aggregation DB

# Rust Service (each node, self-contained)
[mode]
type = "standalone"  # Self-contained, no API dependency

[database]
type = "libsql"  # or "postgresql"
path = "./uppe-node-1.db"  # Local DB per node

[peerup]
enabled = true
# P2P network for peer communication
```

## Recommended Architecture: Decentralized P2P Core

**All Rust services JOIN the decentralized P2P network by default - they strengthen the network, regardless of device type (mobile phones, self-hosters, servers, edge devices):**

```rust
// Rust service is a decentralized peer
// It reads/writes directly to database
// It participates in P2P network
// No API calls needed for core functionality

// Rust service startup logic
let db = open_database(&config.database_path).await?;

// Read monitors directly from DB
let monitors = db.get_enabled_monitors().await?;

// Perform monitoring checks
for monitor in monitors {
    let result = check_monitor(&monitor).await?;
    // Write result directly to DB
    db.save_result(&result).await?;
    
    // Share result with peers via P2P network
    // This strengthens the network
    p2p_network.share_result(&result).await?;
}

// P2P networking (default: join decentralized network)
// All instances join the network by default
if config.peerup.enabled {
    let p2p_network = start_p2p_network().await?;
    
    // Join the decentralized monitoring network
    // Self-hosters, mobile phones, servers - all join the same network
    p2p_network.join_network().await?;
    
    // Listen for results from other peers
    // Each peer strengthens the network (mobile phones = self-hosters = servers = equal)
    p2p_network.on_peer_result(|result| {
        // Store peer results (redundant monitoring)
        db.save_peer_result(&result).await?;
    });
    
    // Share our results with peers
    // All instances contribute to network strength
    p2p_network.share_result(&result).await?;
} else {
    // Optional: Isolated mode (no network participation)
    // Loses decentralized benefits but still works locally
    log::info!("Running in isolated mode (no P2P network)");
}
```

**Go API is optional and independent - only for local frontend:**

```go
// Go API only reads from database for frontend
// It doesn't control Rust service
// It doesn't control the network
// It's purely for local data visibility

// Go API startup logic
db := open_database(&config.database_path)

// Serve frontend API
// - List monitors (read from DB)
// - Get results (read from DB, includes peer results)
// - Aggregations (read from DB)
// - No writes (Rust service handles all writes)
// - No network control (P2P network is autonomous)
```

## Database Access Strategy

### Rust Service (Always Self-Contained)
- **Direct database access** (always)
- **Read monitors** from DB
- **Write results** to DB
- **No API calls** for core functionality
- **Works offline** (no network dependency)

### Go API (Frontend Only)
- **Read-only database access** (recommended)
- **Read monitors** for display
- **Read results** for dashboards
- **Aggregations** for analytics
- **No writes** (Rust service handles all writes)

### Shared Database (Mode 2)
- **SQLite WAL mode** for concurrent reads/writes
- **Rust service**: Read/Write access
- **Go API**: Read-only access (recommended)
- **Connection pooling** in both services

## Performance Optimizations

### Mobile/Standalone
1. **Embedded SQLite** - Zero network overhead
2. **Lightweight HTTP server** - Minimal memory footprint
3. **Batch writes** - Group result writes to reduce I/O
4. **Connection pooling** - Reuse database connections
5. **WAL mode** - Better concurrency for SQLite

### Self-Hosted
1. **Shared database** - Direct access, no HTTP overhead
2. **SQLite WAL mode** - Concurrent reads/writes
3. **Connection pooling** - Both services use pools
4. **Batch inserts** - Group result writes

### Large Scale
1. **PostgreSQL/TimescaleDB** - Better concurrency than SQLite
2. **Connection pooling** - Essential for performance
3. **Read replicas** - Scale reads independently
4. **Batch API calls** - Group monitor fetches
5. **Caching** - Cache monitors in Rust service

## Implementation Plan

### Phase 1: Self-Contained Rust Service (Mobile Support) - PRIORITY

**Rust Service (Core Functionality):**
1. ✅ Direct database access (already implemented)
2. ✅ Read monitors from DB (implement)
3. ✅ Write results to DB (implement)
4. ✅ P2P networking (PeerUP integration)
5. ✅ No API dependency (fully self-contained)

**Benefits:**
- ✅ Mobile phone support
- ✅ Zero external dependencies
- ✅ Works offline
- ✅ Minimal resource usage
- ✅ Fully autonomous

### Phase 2: Optional Go API (Frontend Support)

**Go API (Frontend Only):**
1. ✅ Read-only database access (recommended)
2. ✅ Serve frontend API endpoints
3. ✅ Data aggregations for dashboards
4. ✅ Authentication/authorization
5. ✅ No control over Rust service (independent)

**Shared Database:**
1. Ensure both services can use same database path
2. Coordinate startup so Rust runs migrations first and Go verifies schema compatibility
3. SQLite WAL mode for concurrent access
4. Test concurrent reads/writes

### Phase 3: Distributed Mode (Large Scale)

**Enhancements:**
1. P2P result sharing between Rust services
2. Go API aggregates from multiple nodes
3. Centralized aggregation database (optional)
4. Multi-region support

## Configuration Examples

### Mobile Phone (Joins Network)
```toml
# Rust Service - Joins decentralized network by default
[mode]
type = "standalone"

[database]
type = "libsql"
path = "./uppe.db"

[peerup]
enabled = true  # Default: joins decentralized network
# Optional: set to false for isolated mode (loses network benefits)

[monitoring]
enabled = true
# No API configuration - fully self-contained
```

### Self-Hosted (Joins Network + Local Frontend)
```toml
# Rust Service - Joins decentralized network by default
[mode]
type = "standalone"

[database]
type = "libsql"
path = "/var/lib/uppe/uppe.db"

[peerup]
enabled = true  # Default: joins decentralized network
# Optional: set to false for isolated mode (loses network benefits)

# Go API - Frontend only (optional)
[mode]
type = "api-server"

[database]
type = "libsql"
path = "/var/lib/uppe/uppe.db"  # Same path, read-only recommended

[server]
address = "0.0.0.0:8080"
```

### Isolated Mode (No Network - Optional)
```toml
# Rust Service - Isolated mode (no P2P network)
[mode]
type = "standalone"

[database]
type = "libsql"
path = "./uppe.db"

[peerup]
enabled = false  # Isolated mode - no network participation
# Loses decentralized benefits but still works locally

[monitoring]
enabled = true
```

### Large Scale (Network Itself)
```toml
# All peers (mobile, self-hosted, servers) join this network
# This is the decentralized network configuration
# Each peer uses same P2P network settings

[peerup]
network_id = "uppe-mainnet"  # All peers join same network
bootstrap_nodes = ["..."]  # Network bootstrap nodes
enabled = true  # Default for all peers
```

## Migration Path

1. **Start with Standalone** - Get mobile support working
2. **Add Split Mode** - Enable self-hosted deployments
3. **Add Distributed Mode** - Scale to large deployments

## Resource Usage Estimates

### Mobile Phone (Standalone)
- **Memory**: ~50-100 MB
- **CPU**: < 5% (idle), 10-20% (active monitoring)
- **Battery**: Minimal impact (efficient polling)
- **Storage**: ~10-50 MB (database + binary)

### Self-Hosted (Split)
- **Go API**: ~100-200 MB memory
- **Rust Service**: ~50-100 MB memory
- **Database**: Depends on data size
- **Total**: ~200-400 MB

### Large Scale (Distributed)
- **Go API**: ~500 MB - 2 GB (with caching)
- **Rust Service**: ~100-200 MB per node
- **Database**: Depends on scale (PostgreSQL cluster)

## Recommendations

### For Mobile Phones: **Rust Service Only (Mode 1) - Full P2P Network Peer**
- ✅ **Full P2P network peer** - equal participant, contributes to network strength
- ✅ **Same capabilities as servers** - mobile phones are equal peers
- ✅ Self-contained Rust service
- ✅ Embedded SQLite
- ✅ Zero external dependencies
- ✅ No API needed (fully autonomous)
- ✅ **Full P2P networking** - shares results, receives results, strengthens network
- ✅ Minimal resource usage (~50-100 MB)
- ✅ Works offline (local monitoring)
- ✅ **Network resilience** - mobile phones strengthen the network just like servers

### For Self-Hosters: **Rust Service + Go API (Mode 2) - Joins Network + Local UI**
- ✅ **JOINS decentralized network by default** - participates in large monitoring network
- ✅ **Full P2P network peer** - equal participant, contributes to network strength
- ✅ **Same P2P capabilities as Mode 1** - full network participation
- ✅ **Strengthens network** - contributes monitoring results to decentralized network
- ✅ **Receives network benefits** - gets results from other peers (redundant monitoring)
- ✅ Self-contained Rust service (core functionality)
- ✅ **Full P2P networking** - shares results, receives results, strengthens network
- ✅ Optional Go API (only for local frontend)
- ✅ Optional isolation mode (can disable P2P, loses network benefits)
- ✅ Shared SQLite (small) or PostgreSQL (large)
- ✅ Direct database access (no API overhead)
- ✅ Easy deployment
- ✅ Rust service operates independently

### The Decentralized Network: **All Modes Join (Mode 3) - The Network Itself**
- ✅ **All instances join by default** - mobile phones, self-hosters, servers
- ✅ **All peers are equal** - mobile phones = self-hosters = servers = edge devices
- ✅ P2P network is the core (no central authority)
- ✅ Each peer is autonomous and equal (same capabilities)
- ✅ Network resilience grows with each peer (including mobile phones and self-hosters)
- ✅ Redundant monitoring (multiple peers monitor same URLs)
- ✅ Self-hosters strengthen the network by joining
- ✅ Mobile phones strengthen the network by joining
- ✅ Optional Go API aggregates data for frontend (some peers)
- ✅ PostgreSQL/TimescaleDB for aggregations (optional, some peers)
- ✅ Horizontally scalable (add any device as peer to strengthen network)

## Next Steps

1. **Complete Full P2P Network Peer (Rust Service)** (Priority)
   - ✅ Direct database access (already implemented)
   - Implement monitor reading from DB
   - Implement result writing to DB
   - Integrate PeerUP P2P networking
   - **Full network participation** - share results with peers (strengthen network)
   - **Full network participation** - receive results from peers (redundant monitoring)
   - **Works on all devices** - mobile phones, servers, edge devices (all equal peers)
   - No API dependency (fully autonomous peer)
   - **Mobile-optimized** - efficient P2P networking for mobile devices

2. **Ensure Go API works independently**
   - Read-only database access (recommended)
   - Frontend API endpoints
   - Data aggregations (local + peer results)
   - No control over Rust service or network
   - Optional component (network works without it)

3. **Test Shared Database Access**
   - SQLite WAL mode for concurrency
   - Rust: Read/Write, Go API: Read-only
   - Coordinate migrations
   - Test concurrent access

4. **Enhance P2P Network (All Devices)**
   - Cryptographically signed results
   - Result sharing between peers (all devices equal)
   - Redundant monitoring coordination
   - Network strength metrics
   - Peer discovery and communication
   - **Mobile optimization** - efficient P2P for mobile devices (battery, bandwidth)
   - **Equal participation** - mobile phones have same P2P capabilities as servers

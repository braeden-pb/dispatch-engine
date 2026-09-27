# High-Performance Geospatial Dispatch Engine

A concurrent, low-latency in-memory geospatial dispatch and matching service built in Go for real-time ride-hailing workloads.

## Architecture Overview

```text
dispatch-engine/
├── cmd/
│   ├── server/          # HTTP server entrypoint
│   └── loadgen/         # Multi-threaded concurrent load generator
├── internal/
│   ├── ingest/          # HTTP handlers and request validation
│   ├── match/           # Haversine distance, spatial grid indexing, and expanding-ring search
│   └── store/           # Thread-safe in-memory driver store
└── README.md
```

### Core Components

1. Spatial Grid Indexing
   - Replaces naive O(N) nearest-driver scans with a discretized 2D grid keyed by latitude/longitude buckets.
   - Reduces lookup cost substantially for dense driver distributions.

2. Expanding-Ring Search
   - Starts at the rider's local cell and expands outward in concentric rings when sparse cells are empty.
   - Prevents false 404 responses in suburban or low-density areas.

3. Thread-Safe In-Memory Store
   - Uses Go's sync.RWMutex to allow concurrent reads while preserving write safety for driver location updates.

## Performance

Measured on a simulated fleet of up to 50,000 active drivers:

| Operation                  | Strategy                         | Result                                       |
| -------------------------- | -------------------------------- | -------------------------------------------- |
| Nearest-driver lookup      | Linear scan                      | ~2.79 ms per query                           |
| Nearest-driver lookup      | Spatial grid index               | ~12.75 µs per query                          |
| Concurrent load generation | 50 workers, 70/30 write/read mix | 15,200+ RPS with no transport-level failures |

This demonstrates a roughly 219x improvement in the lookup path under the local benchmark harness.

## Getting Started

### Prerequisites

- Go 1.22+

### Run the server

```powershell
go run ./cmd/server
```

### Run the load generator

```powershell
go run ./cmd/loadgen/main.go -concurrency=50 -duration=10s
```

### Run the test suite

```powershell
go test ./...
```

### Run with race detection

```powershell
go test -race ./...
```

> Note: race detection requires a working 64-bit C toolchain on the host machine.

## API Endpoints

### 1. Ingest a Driver Location

- URL: POST /drivers/location
- Content-Type: application/json

Request body:

```json
{
  "driver_id": "driver-123",
  "lat": 43.6532,
  "lng": -79.3832
}
```

Response:

- 202 Accepted

### 2. Match a Rider to the Nearest Driver

- URL: POST /match
- Content-Type: application/json

Request body:

```json
{
  "rider_lat": 43.6500,
  "rider_lng": -79.3800
}
```

Response:

- 200 OK with the nearest driver object
- 404 Not Found if no driver is available within the search radius

## Summary

This project demonstrates a practical real-time dispatch design using:

- in-memory state
- spatial bucketing
- ring-based fallback recovery
- concurrent load validation

It is a solid foundation for a production-grade geospatial dispatch or ride-matching service.

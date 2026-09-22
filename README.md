# Distributed Geospatial Event Streaming Engine (v0 — skeleton)

## What's here right now
A minimal Go HTTP service that:
- Accepts driver location updates (`POST /drivers/location`)
- Stores them in a thread-safe in-memory map
- Lets you list all current driver locations (`GET /drivers`)

## What's NOT here yet (this is your work)
- `internal/match/match.go` — the actual nearest-driver matching algorithm.
  See the comments in that file for the full task breakdown (Haversine
  distance, complexity, staleness handling).
- A `/match` HTTP endpoint wiring rider requests to `match.FindNearestDriver`.
- Kafka, Redis, PostGIS, Docker — intentionally deferred. Get the naive
  single-process version fully working and understood first. You'll feel
  exactly where it breaks (multiple instances need shared state → Redis;
  ingestion volume spikes → Kafka partitioning; spatial queries at scale →
  PostGIS) before reaching for each piece, instead of cargo-culting them in.

## Run it
```
go build -o bin/server ./cmd/server
./bin/server
```

Then in another terminal:
```
curl -X POST localhost:8080/drivers/location \
  -d '{"driver_id":"d-1","lat":43.65,"lng":-79.38}'

curl localhost:8080/drivers
```

## Suggested next steps, in order
1. Implement `FindNearestDriver` in `internal/match/match.go`. Write the
   Haversine formula yourself. Write a unit test with a known distance
   to check your math.
2. Wire up `POST /match` in `cmd/server/main.go`.
3. Write a load-testing script (even a simple bash loop with `curl`, or
   `hey`/`vegeta`) that posts N driver updates and measures response time.
4. Once the naive O(n) scan is a proven bottleneck under load, add
   geohash-bucketed matching and benchmark old vs. new.
5. Only then: containerize with Docker, add Redis for cross-instance state,
   swap the in-memory store for PostGIS-backed persistence.

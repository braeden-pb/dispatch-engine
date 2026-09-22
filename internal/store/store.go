package store

import (
	"sync"
	"time"
)

// DriverLocation represents a single driver's last known position.
type DriverLocation struct {
	DriverID  string    `json:"driver_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemStore is a thread-safe in-memory store of the latest location per driver.
//
// This is intentionally naive: a single map guarded by a mutex. It works fine
// for one process and a handful of writers/readers. It will NOT work once you
// have multiple service instances (each would have its own copy of the map) —
// that's the exact problem Redis solves later, once you get there. Don't skip
// ahead to Redis until you've felt this limitation yourself.
type MemStore struct {
	mu        sync.RWMutex
	locations map[string]DriverLocation
}

func NewMemStore() *MemStore {
	return &MemStore{
		locations: make(map[string]DriverLocation),
	}
}

// Upsert records or overwrites a driver's latest location.
func (s *MemStore) Upsert(loc DriverLocation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.locations[loc.DriverID] = loc
}

// All returns a snapshot copy of every known driver location.
// Copying under the lock keeps callers from racing with future writes.
func (s *MemStore) All() []DriverLocation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]DriverLocation, 0, len(s.locations))
	for _, loc := range s.locations {
		out = append(out, loc)
	}
	return out
}

// Get returns a single driver's last known location, if present.
func (s *MemStore) Get(driverID string) (DriverLocation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	loc, ok := s.locations[driverID]
	return loc, ok
}

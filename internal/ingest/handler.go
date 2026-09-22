package ingest

import (
	"encoding/json"
	"net/http"
	"time"

	"geo-stream-engine/internal/store"
)

// Handler wires HTTP requests into the store. This layer is deliberately
// "dumb plumbing" — parse JSON, validate, write to store. The interesting
// engineering (matching, concurrency correctness, later: partitioning by
// geohash) happens elsewhere.
type Handler struct {
	Store *store.MemStore
}

func NewHandler(s *store.MemStore) *Handler {
	return &Handler{Store: s}
}

type locationUpdateRequest struct {
	DriverID string  `json:"driver_id"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

// PostLocation handles POST /drivers/location
// Body: {"driver_id": "d-123", "lat": 43.65, "lng": -79.38}
func (h *Handler) PostLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req locationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	// TODO(you): what validation actually matters here? Empty driver_id?
	// Lat/lng out of range (-90..90, -180..180)? Decide and add it —
	// this is the kind of "boring but real" decision interviewers probe.
	if req.DriverID == "" {
		http.Error(w, "driver_id is required", http.StatusBadRequest)
		return
	}

	h.Store.Upsert(store.DriverLocation{
		DriverID:  req.DriverID,
		Lat:       req.Lat,
		Lng:       req.Lng,
		UpdatedAt: time.Now(),
	})

	w.WriteHeader(http.StatusAccepted)
}

// GetDrivers handles GET /drivers — returns every known driver location.
// Mostly here so you can curl something and see state without a DB client.
func (h *Handler) GetDrivers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Store.All())
}

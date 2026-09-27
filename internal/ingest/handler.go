package ingest

import (
	"encoding/json"
	"net/http"
	"time"

	"dispatch-engine/internal/match"
	"dispatch-engine/internal/store"
)

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

type matchRequest struct {
	RiderLat float64 `json:"rider_lat"`
	RiderLng float64 `json:"rider_lng"`
}

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

	if req.DriverID == "" {
		http.Error(w, "driver_id is required", http.StatusBadRequest)
		return
	}

	if req.Lat < -90 || req.Lat > 90 {
		http.Error(w, "lat must be between -90 and 90", http.StatusBadRequest)
		return
	}

	if req.Lng < -180 || req.Lng > 180 {
		http.Error(w, "lng must be between -180 and 180", http.StatusBadRequest)
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

func (h *Handler) GetDrivers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Store.All())
}

func (h *Handler) PostMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req matchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if req.RiderLat < -90 || req.RiderLat > 90 {
		http.Error(w, "rider_lat must be between -90 and 90", http.StatusBadRequest)
		return
	}

	if req.RiderLng < -180 || req.RiderLng > 180 {
		http.Error(w, "rider_lng must be between -180 and 180", http.StatusBadRequest)
		return
	}

	drivers := h.Store.All()
	const (
		gridCellSizeDeg = 0.05
		maxSearchRings  = 3
	)
	index := match.BuildGridIndex(drivers, gridCellSizeDeg)

	nearest, ok := match.FindNearestDriverExpandingRing(req.RiderLat, req.RiderLng, index, gridCellSizeDeg, maxSearchRings)
	if !ok {
		http.Error(w, "no drivers available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nearest)
}

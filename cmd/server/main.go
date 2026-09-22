package main

import (
	"log"
	"net/http"

	"geo-stream-engine/internal/ingest"
	"geo-stream-engine/internal/store"
)

func main() {
	s := store.NewMemStore()
	h := ingest.NewHandler(s)

	mux := http.NewServeMux()
	mux.HandleFunc("/drivers/location", h.PostLocation) // POST: driver sends location update
	mux.HandleFunc("/drivers", h.GetDrivers)             // GET: see all current driver locations

	// TODO(you): add a POST /match endpoint once match.FindNearestDriver
	// is implemented — it should take a rider lat/lng, pull s.All(), call
	// FindNearestDriver, and return the result as JSON. Wire it up the
	// same way PostLocation and GetDrivers are wired above.

	addr := ":8080"
	log.Printf("geo-stream-engine listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

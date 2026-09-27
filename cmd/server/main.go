package main

import (
	"log"
	"net/http"

	"dispatch-engine/internal/ingest"
	"dispatch-engine/internal/store"
)

func main() {
	s := store.NewMemStore()
	h := ingest.NewHandler(s)

	mux := http.NewServeMux()
	mux.HandleFunc("/drivers/location", h.PostLocation) // POST: driver sends location update
	mux.HandleFunc("/drivers", h.GetDrivers)            // GET: see all current driver locations
	mux.HandleFunc("/match", h.PostMatch)               // POST: find nearest driver to a rider

	addr := ":8080"
	log.Printf("dispatch-engine listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

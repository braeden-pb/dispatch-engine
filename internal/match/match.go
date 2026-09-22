package match

import (
	"geo-stream-engine/internal/store"
)

// This file is intentionally unfinished. This is the part of the project
// that's actually yours to design and defend in an interview — don't just
// fill it in mechanically, understand the choices as you make them.
//
// Your task: implement FindNearestDriver, which takes a rider's location
// and a list of candidate driver locations, and returns the closest one.
//
// Things to actually think about (an interviewer will ask about these):
//
//  1. Distance calculation: lat/lng aren't a flat grid — 1 degree of
//     longitude is a different real-world distance depending on latitude.
//     A naive Euclidean distance on raw lat/lng is WRONG. Look up the
//     Haversine formula (great-circle distance) and implement it yourself
//     rather than pulling in a library — you want to be able to explain it.
//
//  2. Complexity: this naive version is O(n) — it scans every driver.
//     Fine for a demo with hundreds of drivers. At real scale (millions of
//     drivers, thousands of ride requests per second) this falls over.
//     This is EXACTLY why real systems (Uber, Lyft) bucket drivers into
//     geospatial cells (H3, S2, or a simple geohash grid) so a match only
//     scans nearby cells instead of the whole world. Once your naive
//     version works, come back and add a geohash-bucketed version, then
//     benchmark the two against each other — that comparison is a great
//     interview talking point.
//
//  3. Staleness: a driver whose last update was 20 minutes ago probably
//     isn't actually there anymore. Should you filter out stale locations?
//     What's a reasonable staleness threshold, and why?
//
// Suggested function signature to start from:
//
//   func FindNearestDriver(riderLat, riderLng float64, drivers []store.DriverLocation) (store.DriverLocation, bool)
//
// Returns the nearest driver and true, or a zero value and false if the
// drivers slice is empty.

func FindNearestDriver(riderLat, riderLng float64, drivers []store.DriverLocation) (store.DriverLocation, bool) {
	// TODO(you): implement this.
	//
	// Rough shape:
	//   1. If len(drivers) == 0, return zero value, false.
	//   2. Track a "best so far" driver and its distance.
	//   3. Loop over drivers, compute haversineDistance(riderLat, riderLng, d.Lat, d.Lng)
	//      for each, keep the minimum.
	//   4. Return the winner.
	//
	// You'll also want a haversineDistance(lat1, lng1, lat2, lng2 float64) float64
	// helper in this same file. Look up the formula, implement it, and write
	// a unit test with two known coordinates and a known distance to check
	// your math (e.g. downtown Toronto to Pearson airport is roughly 27km).
	var zero store.DriverLocation
	return zero, false
}

package match

import (
	"math"
	"testing"

	"dispatch-engine/internal/store"
)

// TestHaversineDistance checks the distance calculation against a pair of
// real-world coordinates with a known, independently-verifiable distance.
// Downtown Toronto to Pearson Airport is commonly cited as ~20-22km by
// road/air, so we assert the result falls in a reasonable window rather
// than pinning an exact float (great-circle "as the crow flies" distance
// will differ slightly from any single published road/flight distance).
func TestHaversineDistance(t *testing.T) {
	const (
		torontoLat = 43.6532
		torontoLng = -79.3832
		pearsonLat = 43.6777
		pearsonLng = -79.6248
	)

	got := haversineDistance(torontoLat, torontoLng, pearsonLat, pearsonLng)

	const wantMin = 18.0
	const wantMax = 24.0
	if got < wantMin || got > wantMax {
		t.Errorf("haversineDistance(Toronto, Pearson) = %.2f km, want between %.1f and %.1f km", got, wantMin, wantMax)
	}
}

// TestHaversineDistance_SamePoint checks the degenerate case: distance from
// a point to itself must be zero. This catches sign errors or bad edge-case
// handling in the trig that a single "normal" test case might miss.
func TestHaversineDistance_SamePoint(t *testing.T) {
	got := haversineDistance(43.65, -79.38, 43.65, -79.38)

	// Floating point trig won't land on an exact 0.0, so we check it's
	// within a tiny epsilon instead of asserting equality outright.
	const epsilon = 0.0001
	if math.Abs(got) > epsilon {
		t.Errorf("haversineDistance(same point) = %.6f km, want ~0", got)
	}
}

// TestFindNearestDriver_ReturnsClosest checks the core matching behavior:
// given a rider location and multiple drivers, the closer driver wins.
func TestFindNearestDriver_ReturnsClosest(t *testing.T) {
	drivers := []store.DriverLocation{
		{DriverID: "far", Lat: 43.70, Lng: -79.40},
		{DriverID: "near", Lat: 43.65, Lng: -79.38},
	}

	riderLat, riderLng := 43.66, -79.39

	got, ok := FindNearestDriver(riderLat, riderLng, drivers)
	if !ok {
		t.Fatal("FindNearestDriver returned ok=false, want ok=true")
	}
	if got.DriverID != "near" {
		t.Errorf("FindNearestDriver returned driver %q, want %q", got.DriverID, "near")
	}
}

// TestFindNearestDriver_EmptySlice checks the guard clause: no drivers
// available should return ok=false, not a zero-value driver mistaken for
// a real match.
func TestFindNearestDriver_EmptySlice(t *testing.T) {
	_, ok := FindNearestDriver(43.66, -79.39, []store.DriverLocation{})
	if ok {
		t.Error("FindNearestDriver with empty slice returned ok=true, want ok=false")
	}
}

// BenchmarkFindNearestDriver measures how FindNearestDriver's O(n) scan
// performs against 10,000 candidate drivers. This is the real, measured
// baseline to compare against once a geohash-bucketed version exists.
func BenchmarkFindNearestDriver(b *testing.B) {
	drivers := make([]store.DriverLocation, 10000)
	for i := range drivers {
		drivers[i] = store.DriverLocation{
			DriverID: "driver",
			Lat:      43.0 + float64(i)*0.0001,
			Lng:      -79.0 + float64(i)*0.0001,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindNearestDriver(43.65, -79.38, drivers)
	}
}

// TestFindNearestDriverGridIndexed_MatchesFullScan verifies the optimized
// grid-indexed lookup agrees with the plain linear scan on the same input.
// A faster wrong answer is worse than a slow correct one — this test is
// what makes the optimization trustworthy.
func TestFindNearestDriverGridIndexed_MatchesFullScan(t *testing.T) {
	drivers := []store.DriverLocation{
		{DriverID: "a", Lat: 43.65, Lng: -79.38},
		{DriverID: "b", Lat: 43.70, Lng: -79.40},
		{DriverID: "c", Lat: 43.66, Lng: -79.39},
	}
	const cellSize = 0.05

	index := BuildGridIndex(drivers, cellSize)

	full, fullOk := FindNearestDriver(43.655, -79.385, drivers)
	grid, gridOk := FindNearestDriverGridIndexed(43.655, -79.385, index, cellSize)

	if fullOk != gridOk {
		t.Fatalf("ok mismatch: full=%v grid=%v", fullOk, gridOk)
	}
	if full.DriverID != grid.DriverID {
		t.Errorf("driver mismatch: full scan picked %q, grid-indexed picked %q", full.DriverID, grid.DriverID)
	}
}

// BenchmarkFindNearestDriverGridIndexed measures lookup cost only — the
// index is built once outside the timer, mirroring how a real system would
// maintain the index incrementally rather than rebuild it per query.
func BenchmarkFindNearestDriverGridIndexed(b *testing.B) {
	drivers := make([]store.DriverLocation, 10000)
	for i := range drivers {
		drivers[i] = store.DriverLocation{
			DriverID: "driver",
			Lat:      43.0 + float64(i)*0.0001,
			Lng:      -79.0 + float64(i)*0.0001,
		}
	}

	const cellSize = 0.05
	index := BuildGridIndex(drivers, cellSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindNearestDriverGridIndexed(43.65, -79.38, index, cellSize)
	}
}

// TestFindNearestDriverExpandingRing_SparseFallback asserts that when the
// immediate 3x3 neighborhood (rings 0 and 1) is completely empty, the search
// successfully expands to outer rings (ring 2+) without returning a false 404.
func TestFindNearestDriverExpandingRing_SparseFallback(t *testing.T) {
	const cellSize = 0.05
	riderLat, riderLng := 43.65, -79.38

	// Place a driver exactly 2 cells away in latitude (~11 km north)
	// Outer ring k=2: distance is 2 * cellSize = 0.10 degrees offset
	farDriver := store.DriverLocation{
		DriverID: "suburban-driver",
		Lat:      riderLat + 2.0*cellSize,
		Lng:      riderLng,
	}

	drivers := []store.DriverLocation{farDriver}
	index := BuildGridIndex(drivers, cellSize)

	// 1. Fixed 3x3 scan (k=1) must fail to locate the driver
	_, ok3x3 := FindNearestDriverGridIndexed(riderLat, riderLng, index, cellSize)
	if ok3x3 {
		t.Fatal("FindNearestDriverGridIndexed found driver unexpectedly in 3x3 neighborhood")
	}

	// 2. Expanding ring up to maxRings=1 must also fail
	_, okRing1 := FindNearestDriverExpandingRing(riderLat, riderLng, index, cellSize, 1)
	if okRing1 {
		t.Fatal("FindNearestDriverExpandingRing found driver when maxRings was restricted to 1")
	}

	// 3. Expanding ring up to maxRings=2 or higher must successfully locate the driver
	matchedDriver, okRing2 := FindNearestDriverExpandingRing(riderLat, riderLng, index, cellSize, 2)
	if !okRing2 {
		t.Fatalf("FindNearestDriverExpandingRing failed to find driver in ring 2, got ok=false")
	}

	if matchedDriver.DriverID != "suburban-driver" {
		t.Errorf("matched driver = %q, want %q", matchedDriver.DriverID, "suburban-driver")
	}
}

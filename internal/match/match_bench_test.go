package match_test

import (
	"math/rand"
	"testing"
	"time"

	"dispatch-engine/internal/match"
	"dispatch-engine/internal/store"
)

var (
	drivers1K  = generateDrivers(1000)
	drivers10K = generateDrivers(10000)
	drivers50K = generateDrivers(50000)

	grid1K  = match.BuildGridIndex(drivers1K, 0.05)
	grid10K = match.BuildGridIndex(drivers10K, 0.05)
	grid50K = match.BuildGridIndex(drivers50K, 0.05)
)

func generateDrivers(n int) []store.DriverLocation {
	rng := rand.New(rand.NewSource(42))
	drivers := make([]store.DriverLocation, n)
	for i := 0; i < n; i++ {
		drivers[i] = store.DriverLocation{
			DriverID:  "driver",
			Lat:       43.60 + rng.Float64()*0.20,
			Lng:       -79.50 + rng.Float64()*0.30,
			UpdatedAt: time.Now(),
		}
	}
	return drivers
}

const (
	riderLat    = 43.6532
	riderLng    = -79.3832
	cellSizeDeg = 0.05
)

// Linear scans: O(N)
func BenchmarkLinear_1000(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = match.FindNearestDriver(riderLat, riderLng, drivers1K)
	}
}

func BenchmarkLinear_10000(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = match.FindNearestDriver(riderLat, riderLng, drivers10K)
	}
}

func BenchmarkLinear_50000(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = match.FindNearestDriver(riderLat, riderLng, drivers50K)
	}
}

// Spatial Grid Lookups: O(1)
func BenchmarkGrid_1000(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = match.FindNearestDriverGridIndexed(riderLat, riderLng, grid1K, cellSizeDeg)
	}
}

func BenchmarkGrid_10000(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = match.FindNearestDriverGridIndexed(riderLat, riderLng, grid10K, cellSizeDeg)
	}
}

func BenchmarkGrid_50000(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = match.FindNearestDriverGridIndexed(riderLat, riderLng, grid50K, cellSizeDeg)
	}
}

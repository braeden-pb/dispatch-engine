package match

import (
	"log"
	"math"

	"dispatch-engine/internal/store"
)

// toRadians converts degrees to radians (Go's math funcs expect radians).
func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

// haversineDistance returns great-circle distance in km between two points.
// Plain Euclidean distance on lat/lng is wrong because longitude lines
// converge toward the poles, this accounts for that.
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := toRadians(lat1)
	lat2Rad := toRadians(lat2)
	deltaLat := toRadians(lat2 - lat1)
	deltaLng := toRadians(lng2 - lng1)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a)) // atan2 for numerical stability

	return earthRadiusKm * c
}

// FindNearestDriver returns the closest driver to the rider's location.
func FindNearestDriver(riderLat, riderLng float64, drivers []store.DriverLocation) (store.DriverLocation, bool) {
	var zero store.DriverLocation

	if len(drivers) == 0 {
		log.Printf("FindNearestDriver: empty slice, returning false") // TEMP DEBUG
		return zero, false
	}
	log.Printf("FindNearestDriver: got %d drivers, proceeding", len(drivers)) // TEMP DEBUG

	nearest := drivers[0]
	minDistance := haversineDistance(riderLat, riderLng, nearest.Lat, nearest.Lng)

	for _, driver := range drivers[1:] {
		distance := haversineDistance(riderLat, riderLng, driver.Lat, driver.Lng)
		if distance < minDistance {
			minDistance = distance
			nearest = driver
		}
	}

	return nearest, true
}

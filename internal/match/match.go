package match

import (
	"fmt"
	"math"

	"dispatch-engine/internal/store"
)

// GridKey buckets a lat/lng pair into a grid cell identifier.
func GridKey(lat, lng, cellSizeDeg float64) string {
	latIdx := int64(math.Floor(lat / cellSizeDeg))
	lngIdx := int64(math.Floor(lng / cellSizeDeg))
	return fmt.Sprintf("%d:%d", latIdx, lngIdx)
}

// NeighborKeys returns the center grid cell and all 8 immediate adjacent cells.
func NeighborKeys(lat, lng, cellSizeDeg float64) []string {
	riderLatIdx := int64(math.Floor(lat / cellSizeDeg))
	riderLngIdx := int64(math.Floor(lng / cellSizeDeg))

	keys := make([]string, 0, 9)
	for di := int64(-1); di <= 1; di++ {
		for dj := int64(-1); dj <= 1; dj++ {
			keys = append(keys, fmt.Sprintf("%d:%d", riderLatIdx+di, riderLngIdx+dj))
		}
	}
	return keys
}

// BuildGridIndex buckets every driver into its grid cell.
func BuildGridIndex(drivers []store.DriverLocation, cellSizeDeg float64) map[string][]store.DriverLocation {
	index := make(map[string][]store.DriverLocation)
	for _, d := range drivers {
		key := GridKey(d.Lat, d.Lng, cellSizeDeg)
		index[key] = append(index[key], d)
	}
	return index
}

// FindNearestDriverGridIndexed finds the nearest driver by scanning neighbor cells.
func FindNearestDriverGridIndexed(riderLat, riderLng float64, index map[string][]store.DriverLocation, cellSizeDeg float64) (store.DriverLocation, bool) {
	var candidates []store.DriverLocation
	for _, key := range NeighborKeys(riderLat, riderLng, cellSizeDeg) {
		candidates = append(candidates, index[key]...)
	}

	if len(candidates) == 0 {
		var zero store.DriverLocation
		return zero, false
	}

	return FindNearestDriver(riderLat, riderLng, candidates)
}

// toRadians converts degrees to radians (Go's math funcs expect radians).
func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

// haversineDistance returns great-circle distance in km between two points.
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := toRadians(lat1)
	lat2Rad := toRadians(lat2)
	deltaLat := toRadians(lat2 - lat1)
	deltaLng := toRadians(lng2 - lng1)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

// FindNearestDriver returns the closest driver to the rider's location.
func FindNearestDriver(riderLat, riderLng float64, drivers []store.DriverLocation) (store.DriverLocation, bool) {
	var zero store.DriverLocation
	if len(drivers) == 0 {
		return zero, false
	}

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

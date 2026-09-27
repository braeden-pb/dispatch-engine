package match

import (
	"fmt"
	"math"

	"dispatch-engine/internal/store"
)

func gridIndex(value, cellSizeDeg float64) int64 {
	const epsilon = 1e-9
	return int64(math.Floor((value + epsilon) / cellSizeDeg))
}

// GridKey buckets a lat/lng pair into a grid cell identifier.
func GridKey(lat, lng, cellSizeDeg float64) string {
	latIdx := gridIndex(lat, cellSizeDeg)
	lngIdx := gridIndex(lng, cellSizeDeg)
	return fmt.Sprintf("%d:%d", latIdx, lngIdx)
}

// NeighborKeys returns the center grid cell and all 8 immediate adjacent cells.
func NeighborKeys(lat, lng, cellSizeDeg float64) []string {
	riderLatIdx := gridIndex(lat, cellSizeDeg)
	riderLngIdx := gridIndex(lng, cellSizeDeg)

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

// FindNearestDriverGridIndexed finds the nearest driver by scanning only the
// rider's current grid cell and its immediate neighbors, without allocating a
// temporary candidate slice for the whole neighborhood.
func FindNearestDriverGridIndexed(riderLat, riderLng float64, index map[string][]store.DriverLocation, cellSizeDeg float64) (store.DriverLocation, bool) {
	var nearest store.DriverLocation
	var found bool
	var minDistance float64

	for _, key := range NeighborKeys(riderLat, riderLng, cellSizeDeg) {
		for _, driver := range index[key] {
			distance := haversineDistance(riderLat, riderLng, driver.Lat, driver.Lng)
			if !found || distance < minDistance {
				nearest = driver
				minDistance = distance
				found = true
			}
		}
	}

	if !found {
		var zero store.DriverLocation
		return zero, false
	}

	return nearest, true
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

// RingKeys returns the cell keys forming the outer perimeter at Chebyshev distance k.
// k = 0: [center] (1 cell)
// k = 1: 3x3 perimeter excluding center (8 cells)
// k = 2: 5x5 perimeter excluding 3x3 (16 cells)
func RingKeys(lat, lng, cellSizeDeg float64, k int64) []string {
	if k < 0 {
		return nil
	}
	if k == 0 {
		return []string{GridKey(lat, lng, cellSizeDeg)}
	}

	latIdx := gridIndex(lat, cellSizeDeg)
	lngIdx := gridIndex(lng, cellSizeDeg)

	keys := make([]string, 0, 8*k)
	for di := -k; di <= k; di++ {
		for dj := -k; dj <= k; dj++ {
			if int64(math.Abs(float64(di))) == k || int64(math.Abs(float64(dj))) == k {
				keys = append(keys, fmt.Sprintf("%d:%d", latIdx+di, lngIdx+dj))
			}
		}
	}
	return keys
}

// FindNearestDriverExpandingRing searches outward ring-by-ring up to maxRings.
// Dense lookups terminate at ring 0 or 1, while sparse lookups expand gracefully.
func FindNearestDriverExpandingRing(
	riderLat, riderLng float64,
	index map[string][]store.DriverLocation,
	cellSizeDeg float64,
	maxRings int64,
) (store.DriverLocation, bool) {
	var nearest store.DriverLocation
	var found bool
	var minDistance float64

	for ring := int64(0); ring <= maxRings; ring++ {
		for _, key := range RingKeys(riderLat, riderLng, cellSizeDeg, ring) {
			for _, driver := range index[key] {
				dist := haversineDistance(riderLat, riderLng, driver.Lat, driver.Lng)
				if !found || dist < minDistance {
					nearest = driver
					minDistance = dist
					found = true
				}
			}
		}

		// Early exit: once a driver is discovered in the inner ring, return immediately
		if found {
			return nearest, true
		}
	}

	var zero store.DriverLocation
	return zero, false
}

package match

import (
	"fmt"
	"math"

	"dispatch-engine/internal/store"
)

// gridKey buckets a lat/lng pair into a grid cell identifier. cellSizeDeg
// controls the cell size — e.g. 0.05 degrees is roughly 5.5km at the
// equator (longitude cell width shrinks toward the poles, same caveat
// as everywhere else lat/lng is treated as a flat grid).
func gridKey(lat, lng, cellSizeDeg float64) string {
	latIdx := int64(math.Floor(lat / cellSizeDeg))
	lngIdx := int64(math.Floor(lng / cellSizeDeg))
	return fmt.Sprintf("%d:%d", latIdx, lngIdx)
}

// BuildGridIndex buckets every driver into its grid cell. In a real system
// this index would be maintained incrementally as drivers move (update one
// driver's bucket on each location ping) rather than rebuilt from scratch
// on every query — rebuilding per-query here is just for this comparison.
func BuildGridIndex(drivers []store.DriverLocation, cellSizeDeg float64) map[string][]store.DriverLocation {
	index := make(map[string][]store.DriverLocation)
	for _, d := range drivers {
		key := gridKey(d.Lat, d.Lng, cellSizeDeg)
		index[key] = append(index[key], d)
	}
	return index
}

// FindNearestDriverGridIndexed finds the nearest driver by only scanning
// the rider's grid cell and its 8 immediate neighbors, instead of every
// driver in the system.
//
// Known limitation: if the 3x3 neighborhood has zero drivers (sparse area,
// or cellSizeDeg too small for current driver density), this returns
// ok=false rather than expanding the search outward. A production system
// would grow the search radius ring by ring until it finds candidates —
// left out here for simplicity, but worth naming as the honest gap.
func FindNearestDriverGridIndexed(riderLat, riderLng float64, index map[string][]store.DriverLocation, cellSizeDeg float64) (store.DriverLocation, bool) {
	riderLatIdx := int64(math.Floor(riderLat / cellSizeDeg))
	riderLngIdx := int64(math.Floor(riderLng / cellSizeDeg))

	var candidates []store.DriverLocation
	for di := int64(-1); di <= 1; di++ {
		for dj := int64(-1); dj <= 1; dj++ {
			key := fmt.Sprintf("%d:%d", riderLatIdx+di, riderLngIdx+dj)
			candidates = append(candidates, index[key]...)
		}
	}

	if len(candidates) == 0 {
		var zero store.DriverLocation
		return zero, false
	}

	// Reuse the existing linear-scan logic, but now against a much
	// smaller candidate list instead of every driver in the system.
	return FindNearestDriver(riderLat, riderLng, candidates)
}

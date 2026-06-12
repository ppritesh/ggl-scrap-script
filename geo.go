package main

import (
	"math"
	"strconv"
)

const defaultRadiusKm = 15

// haversineKm returns the distance in km between two lat/lng points.
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371

	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	return earthRadiusKm * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func parseCoord(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

// zoomForRadiusKm picks a Maps zoom level that roughly frames the given radius.
func zoomForRadiusKm(radiusKm float64) int {
	switch {
	case radiusKm <= 5:
		return 13
	case radiusKm <= 10:
		return 12
	case radiusKm <= 20:
		return 11
	case radiusKm <= 40:
		return 10
	default:
		return 9
	}
}

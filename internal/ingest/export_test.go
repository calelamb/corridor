package ingest

// InYosemiteCorridorBbox exports inYosemiteCorridorBbox for external package testing.
func InYosemiteCorridorBbox(lon, lat float64) bool {
	return inYosemiteCorridorBbox(lon, lat)
}

// InYellowstoneCorridorBbox exports inYellowstoneCorridorBbox for external package testing.
func InYellowstoneCorridorBbox(lon, lat float64) bool {
	return inYellowstoneCorridorBbox(lon, lat)
}

// InCorridorBbox exports inCorridorBbox for external package testing.
func InCorridorBbox(lon, lat float64) bool {
	return inCorridorBbox(lon, lat)
}

// MatchHighwayRef exports matchHighwayRef for external package testing.
func MatchHighwayRef(tags map[string]string) (string, bool) {
	return matchHighwayRef(tags)
}

// DetermineState exports determineState for external package testing.
func DetermineState(tags map[string]string, lat, lon float64) string {
	return determineState(tags, lat, lon)
}

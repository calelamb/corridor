package api

import (
	"errors"
	"math"
	"strings"
	"time"

	"corridor/internal/auth"

	h3 "github.com/akhenakh/goh3"
)

// publicH3Resolution is the finest Uber H3 resolution allowed on public
// responses. Coarser cells (smaller resolution numbers) also satisfy the policy.
const publicH3Resolution = 6

var sensitiveSpeciesNames = map[string]struct{}{}

var nonSensitiveSpeciesNames = map[string]struct{}{}

func init() {
	for _, name := range []string{
		"sierra nevada bighorn sheep",
		"sierra bighorn sheep",
		"sierra nevada bighorn",
		"bighorn sheep",
		"ovis canadensis sierrae",
		"pacific fisher",
		"fisher",
		"pekania pennanti",
		"martes pennanti",
		"california black bear",
		"black bear",
		"american black bear",
		"ursus americanus",
		"grizzly bear",
		"grizzly",
		"brown bear",
		"ursus arctos",
		"ursus arctos horribilis",
		"canada lynx",
		"lynx",
		"lynx canadensis",
		"gray wolf",
		"grey wolf",
		"wolf",
		"timber wolf",
		"canis lupus",
		"wolverine",
		"gulo gulo",
		"gulo gulo luscus",
	} {
		sensitiveSpeciesNames[normalizeSpeciesName(name)] = struct{}{}
	}
	for _, name := range []string{
		"mule deer",
		"black tailed deer",
		"white tailed deer",
		"elk",
		"wapiti",
		"moose",
		"pronghorn",
		"pronghorn antelope",
		"bison",
		"coyote",
		"bobcat",
		"lynx rufus",
		"mountain lion",
		"cougar",
		"puma",
		"red fox",
	} {
		nonSensitiveSpeciesNames[normalizeSpeciesName(name)] = struct{}{}
	}
}

func normalizeSpeciesName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	return strings.Join(strings.Fields(name), " ")
}

func speciesNameParts(name string) []string {
	norm := normalizeSpeciesName(name)
	if norm == "" {
		return nil
	}
	parts := []string{norm}
	if i := strings.Index(norm, "("); i > 0 {
		parts = append(parts, strings.TrimSpace(norm[:i]))
		if j := strings.Index(norm, ")"); j > i {
			parts = append(parts, strings.TrimSpace(norm[i+1:j]))
		}
	}
	for _, piece := range strings.Split(norm, "/") {
		piece = strings.TrimSpace(piece)
		if piece != "" {
			parts = append(parts, piece)
		}
	}
	return parts
}

func containsSpeciesPhrase(haystack, phrase string) bool {
	if phrase == "" || haystack == "" {
		return false
	}
	return strings.Contains(" "+haystack+" ", " "+phrase+" ")
}

// isSensitiveSpecies reports whether a reported common name is on the
// sensitive-species list (grizzly, lynx, wolf, wolverine, fisher, bighorn, black bear).
func isSensitiveSpecies(name string) bool {
	parts := speciesNameParts(name)
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if _, ok := sensitiveSpeciesNames[part]; ok {
			return true
		}
	}
	for _, part := range parts {
		if _, ok := nonSensitiveSpeciesNames[part]; ok {
			return false
		}
	}
	norm := parts[0]
	for alias := range sensitiveSpeciesNames {
		if containsSpeciesPhrase(norm, alias) {
			return true
		}
	}
	return false
}

// generalizeToH3 snaps a point to the centroid of its H3 cell at maxRes,
// never finer than resolution 6.
func generalizeToH3(lat, lon float64, maxRes int) (float64, float64, error) {
	if math.IsNaN(lat) || math.IsNaN(lon) || math.IsInf(lat, 0) || math.IsInf(lon, 0) {
		return 0, 0, errors.New("invalid coordinate")
	}
	res := maxRes
	if res <= 0 || res > publicH3Resolution {
		res = publicH3Resolution
	}
	cell := h3.LatLngToCell(h3.NewLatLng(lat, lon), res)
	if !cell.IsValid() {
		return 0, 0, errors.New("invalid h3 cell")
	}
	centroid := cell.LatLng()
	return centroid.Lat, centroid.Lng, nil
}

func generalizePoint(lat, lon float64, maxRes int) (float64, float64, bool) {
	gLat, gLon, err := generalizeToH3(lat, lon, maxRes)
	if err != nil {
		return 0, 0, false
	}
	return gLat, gLon, true
}

// redactFieldReport applies spatial generalization for callers who are not
// analysts or admins. Storage stays precise; only the serialized copy changes.
func redactFieldReport(role auth.Role, report FieldReport) (FieldReport, error) {
	filter := auth.SensitiveFilterForRole(role)
	if filter.AllowRawCoordinates || !isSensitiveSpecies(report.SpeciesCommonName) {
		return report, nil
	}
	lat, lon, ok := generalizePoint(report.Latitude, report.Longitude, filter.MaxH3Resolution)
	if !ok {
		return FieldReport{}, errors.New("sensitive coordinate generalization failed")
	}
	out := report
	out.Latitude = lat
	out.Longitude = lon
	if out.SnappedLatitude != nil && out.SnappedLongitude != nil {
		sLat, sLon, snappedOK := generalizePoint(*out.SnappedLatitude, *out.SnappedLongitude, filter.MaxH3Resolution)
		if !snappedOK {
			out.SnappedLatitude = nil
			out.SnappedLongitude = nil
			out.SnapDistanceM = nil
		} else {
			out.SnappedLatitude = &sLat
			out.SnappedLongitude = &sLon
		}
	}
	return out, nil
}

// reportVisibleToRole is false when a sensitive report is still inside the
// public 30-day quarantine. Analysts and admins are never delayed.
func reportVisibleToRole(role auth.Role, report FieldReport, now time.Time) bool {
	filter := auth.SensitiveFilterForRole(role)
	if filter.AllowRawCoordinates || filter.MinAgeDelay <= 0 || !isSensitiveSpecies(report.SpeciesCommonName) {
		return true
	}
	observed := report.ObservedAt
	if observed.IsZero() {
		observed = report.CreatedAt
	}
	if observed.IsZero() {
		return false
	}
	return now.Sub(observed) >= filter.MinAgeDelay
}

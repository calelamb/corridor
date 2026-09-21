package explore

import (
	"errors"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const Policy = "h3-r6-month-v1"

type Filters struct {
	BBox                                   [4]float64
	Species, Source, Road, Season, Cell, Q string
	Start, End, Limit, Offset              int
}

var cellPattern = regexp.MustCompile(`^[0-9a-f]{15}$`)

func Parse(v url.Values) (Filters, error) {
	f := Filters{BBox: [4]float64{-180, -85, 180, 85}, Start: 1900, End: 2026, Limit: 100}
	allowed := map[string]bool{"bbox": true, "species": true, "source": true, "road": true, "season": true, "cell": true, "q": true, "start": true, "end": true, "limit": true, "offset": true}
	for key, values := range v {
		if !allowed[key] || len(values) != 1 || len(values[0]) > 300 || strings.ContainsAny(values[0], "\x00\r\n") {
			return f, errors.New("invalid filter")
		}
	}
	if s := v.Get("bbox"); s != "" {
		parts := strings.Split(s, ",")
		if len(parts) != 4 {
			return f, errors.New("invalid bounds")
		}
		for i, s := range parts {
			n, err := strconv.ParseFloat(s, 64)
			if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
				return f, errors.New("invalid bounds")
			}
			f.BBox[i] = n
		}
	}
	b := f.BBox
	if b[0] < -180 || b[2] > 180 || b[1] < -85 || b[3] > 85 || b[0] >= b[2] || b[1] >= b[3] {
		return f, errors.New("invalid bounds")
	}
	for _, p := range []struct {
		name     string
		dest     *int
		min, max int
	}{{"start", &f.Start, 1900, 2026}, {"end", &f.End, 1900, 2026}, {"limit", &f.Limit, 1, 500}, {"offset", &f.Offset, 0, 100000}} {
		if s := v.Get(p.name); s != "" {
			n, err := strconv.Atoi(s)
			if err != nil || n < p.min || n > p.max {
				return f, errors.New("invalid range")
			}
			*p.dest = n
		}
	}
	if f.Start > f.End {
		return f, errors.New("reversed dates")
	}
	f.Species = v.Get("species")
	f.Source = v.Get("source")
	f.Road = v.Get("road")
	f.Cell = v.Get("cell")
	f.Season = v.Get("season")
	f.Q = v.Get("q")
	if f.Cell != "" && !cellPattern.MatchString(f.Cell) {
		return f, errors.New("invalid cell")
	}
	if f.Season != "" && f.Season != "spring" && f.Season != "summer" && f.Season != "autumn" && f.Season != "winter" {
		return f, errors.New("invalid season")
	}
	return f, nil
}
func (f Filters) args() []any {
	return []any{f.BBox[0], f.BBox[1], f.BBox[2], f.BBox[3], f.Species, f.Source, f.Road, f.Start, f.End, f.Season, f.Cell, f.Q}
}

const predicate = ` WHERE geom && ST_MakeEnvelope($1,$2,$3,$4,4326)
 AND ($5='' OR species=$5) AND ($6='' OR source_id=$6) AND ($7='' OR road=$7)
 AND year BETWEEN $8 AND $9 AND ($10='' OR CASE WHEN month IN(12,1,2) THEN 'winter' WHEN month BETWEEN 3 AND 5 THEN 'spring' WHEN month BETWEEN 6 AND 8 THEN 'summer' WHEN month BETWEEN 9 AND 11 THEN 'autumn' ELSE 'unknown' END=$10)
 AND ($11='' OR cell=$11) AND ($12='' OR strpos(lower(road || ' ' || species),lower($12))>0) `

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedSegments populates benchmark 0.1-mile road segments and risk scores
// for pilot verification and frontend visualization across 5 National Park corridors:
// - Yosemite National Park (CA): CA-120 Crane Flat / Tioga Road MP 40-75 (black bear, mule deer, sierra bighorn)
// - Yellowstone National Park (WY/MT): US-191 Gallatin Canyon MP 0-30 (elk, moose, grizzly bear)
// - Grand Teton National Park (WY): US-89 / US-26 Jackson Hole MP 10-45 (mule deer, pronghorn, elk)
// - Glacier National Park (MT): US-2 Southern Boundary MP 140-185 (grizzly bear, canada lynx, bighorn sheep)
// - Rocky Mountain National Park (CO): US-34 Trail Ridge Road MP 50-70 (elk, mule deer, bighorn sheep)
func SeedSegments(ctx context.Context, pool *pgxpool.Pool, stdout io.Writer) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	if pool == nil {
		return errors.New("database pool is required")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const seedSQL = `
	-- Benchmark road segments across 5 National Park corridors
	INSERT INTO road_segments (
		id, state, highway_ref, milepost_start, milepost_end, length_m, native_id, geom
	) VALUES 
	-- 1. Yellowstone National Park (WY): US-191 Gallatin Canyon MP 0-30
	(
		'00000000-0000-0000-0000-000000000001',
		'WY',
		'US-191',
		14.2,
		14.3,
		160.9344,
		'benchmark-yell-us191-0001',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-111.16 44.85, -111.16 44.86)'), 4326)
	),
	(
		'00000000-0000-0000-0001-000000000001',
		'WY',
		'US-191',
		12.0,
		12.1,
		160.9344,
		'benchmark-yell-us191-0002',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-111.13 44.82, -111.13 44.83)'), 4326)
	),
	(
		'00000000-0000-0000-0001-000000000002',
		'WY',
		'US-191',
		12.1,
		12.2,
		160.9344,
		'benchmark-yell-us191-0003',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-111.13 44.83, -111.13 44.84)'), 4326)
	),
	(
		'00000000-0000-0000-0001-000000000003',
		'WY',
		'US-191',
		12.2,
		12.3,
		160.9344,
		'benchmark-yell-us191-0004',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-111.13 44.84, -111.13 44.85)'), 4326)
	),
	(
		'00000000-0000-0000-0001-000000000004',
		'WY',
		'US-191',
		25.0,
		25.1,
		160.9344,
		'benchmark-yell-us191-0005',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-111.20 45.00, -111.20 45.01)'), 4326)
	),

	-- 2. Grand Teton National Park (WY): US-89 / US-26 Jackson Hole MP 10-45
	(
		'00000000-0000-0000-0002-000000000001',
		'WY',
		'US-89',
		18.0,
		18.1,
		160.9344,
		'benchmark-grte-us89-0001',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-110.66 43.72, -110.66 43.73)'), 4326)
	),
	(
		'00000000-0000-0000-0002-000000000002',
		'WY',
		'US-89',
		18.1,
		18.2,
		160.9344,
		'benchmark-grte-us89-0002',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-110.66 43.73, -110.66 43.74)'), 4326)
	),
	(
		'00000000-0000-0000-0002-000000000003',
		'WY',
		'US-89',
		18.2,
		18.3,
		160.9344,
		'benchmark-grte-us89-0003',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-110.66 43.74, -110.66 43.75)'), 4326)
	),
	(
		'00000000-0000-0000-0002-000000000004',
		'WY',
		'US-89',
		25.0,
		25.1,
		160.9344,
		'benchmark-grte-us89-0004',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-110.58 43.78, -110.58 43.79)'), 4326)
	),
	(
		'00000000-0000-0000-0002-000000000005',
		'WY',
		'US-89',
		35.0,
		35.1,
		160.9344,
		'benchmark-grte-us89-0005',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-110.55 43.84, -110.55 43.85)'), 4326)
	),

	-- 3. Glacier National Park (MT): US-2 Southern Boundary MP 140-185
	(
		'00000000-0000-0000-0003-000000000001',
		'MT',
		'US-2',
		152.0,
		152.1,
		160.9344,
		'benchmark-glac-us2-0001',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-113.62 48.33, -113.61 48.33)'), 4326)
	),
	(
		'00000000-0000-0000-0003-000000000002',
		'MT',
		'US-2',
		152.1,
		152.2,
		160.9344,
		'benchmark-glac-us2-0002',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-113.61 48.33, -113.60 48.33)'), 4326)
	),
	(
		'00000000-0000-0000-0003-000000000003',
		'MT',
		'US-2',
		152.2,
		152.3,
		160.9344,
		'benchmark-glac-us2-0003',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-113.60 48.33, -113.59 48.33)'), 4326)
	),
	(
		'00000000-0000-0000-0003-000000000004',
		'MT',
		'US-2',
		165.0,
		165.1,
		160.9344,
		'benchmark-glac-us2-0004',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-113.35 48.31, -113.34 48.31)'), 4326)
	),
	(
		'00000000-0000-0000-0003-000000000005',
		'MT',
		'US-2',
		180.0,
		180.1,
		160.9344,
		'benchmark-glac-us2-0005',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-113.24 48.42, -113.23 48.43)'), 4326)
	),

	-- 4. Rocky Mountain National Park (CO): US-34 Trail Ridge Road MP 50-70
	(
		'00000000-0000-0000-0004-000000000001',
		'CO',
		'US-34',
		58.0,
		58.1,
		160.9344,
		'benchmark-romo-us34-0001',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-105.77 40.42, -105.76 40.42)'), 4326)
	),
	(
		'00000000-0000-0000-0004-000000000002',
		'CO',
		'US-34',
		58.1,
		58.2,
		160.9344,
		'benchmark-romo-us34-0002',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-105.76 40.42, -105.75 40.42)'), 4326)
	),
	(
		'00000000-0000-0000-0004-000000000003',
		'CO',
		'US-34',
		58.2,
		58.3,
		160.9344,
		'benchmark-romo-us34-0003',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-105.75 40.42, -105.74 40.42)'), 4326)
	),
	(
		'00000000-0000-0000-0004-000000000004',
		'CO',
		'US-34',
		65.0,
		65.1,
		160.9344,
		'benchmark-romo-us34-0004',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-105.75 40.44, -105.74 40.44)'), 4326)
	),
	(
		'00000000-0000-0000-0004-000000000005',
		'CO',
		'US-34',
		69.5,
		69.6,
		160.9344,
		'benchmark-romo-us34-0005',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-105.62 40.38, -105.61 40.38)'), 4326)
	),

	-- 5. Yosemite National Park (CA): CA-120 Crane Flat / Tioga Road MP 40-75 & CA-140
	(
		'00000000-0000-0000-0000-000000000005',
		'CA',
		'CA-120',
		42.0,
		42.1,
		160.9344,
		'benchmark-yose-ca120-0001',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-119.80 37.85, -119.78 37.85)'), 4326)
	),
	(
		'00000000-0000-0000-0005-000000000002',
		'CA',
		'CA-120',
		42.1,
		42.2,
		160.9344,
		'benchmark-yose-ca120-0002',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-119.78 37.85, -119.76 37.85)'), 4326)
	),
	(
		'00000000-0000-0000-0005-000000000003',
		'CA',
		'CA-120',
		42.2,
		42.3,
		160.9344,
		'benchmark-yose-ca120-0003',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-119.76 37.85, -119.74 37.85)'), 4326)
	),
	(
		'00000000-0000-0000-0005-000000000004',
		'CA',
		'CA-120',
		65.0,
		65.1,
		160.9344,
		'benchmark-yose-ca120-0004',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-119.38 37.87, -119.36 37.87)'), 4326)
	),
	(
		'00000000-0000-0000-0005-000000000005',
		'CA',
		'CA-120',
		74.9,
		75.0,
		160.9344,
		'benchmark-yose-ca120-0005',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-119.27 37.91, -119.25 37.91)'), 4326)
	),
	(
		'00000000-0000-0000-0000-000000000006',
		'CA',
		'CA-140',
		8.5,
		8.6,
		160.9344,
		'benchmark-yose-ca140-0006',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-119.75 37.70, -119.73 37.70)'), 4326)
	),

	-- Additional inter-state benchmark corridors (NV, UT, ID)
	(
		'00000000-0000-0000-0000-000000000002',
		'NV',
		'I-80',
		352.1,
		352.2,
		160.9344,
		'benchmark-i80-0002',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-114.70 41.1, -114.68 41.1)'), 4326)
	),
	(
		'00000000-0000-0000-0000-000000000003',
		'UT',
		'US-89',
		45.0,
		45.1,
		160.9344,
		'benchmark-us89-0003',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-111.50 41.8, -111.48 41.8)'), 4326)
	),
	(
		'00000000-0000-0000-0000-000000000004',
		'ID',
		'I-84',
		115.0,
		115.1,
		160.9344,
		'benchmark-i84-0004',
		ST_SetSRID(ST_GeomFromText('LINESTRING(-115.10 43.1, -115.08 43.1)'), 4326)
	)
	ON CONFLICT (id) DO UPDATE SET
		geom = EXCLUDED.geom,
		state = EXCLUDED.state,
		highway_ref = EXCLUDED.highway_ref,
		milepost_start = EXCLUDED.milepost_start,
		milepost_end = EXCLUDED.milepost_end,
		length_m = EXCLUDED.length_m;

	-- Benchmark segment risk scores
	INSERT INTO segment_risk_scores (
		segment_id, risk_score, percentile_rank, p10, p90, monthly_profile, species_breakdown, shap_drivers
	) VALUES
	-- Yellowstone (US-191 MP 12.0-12.3 cluster + MP 14.2 & 25.0)
	(
		'00000000-0000-0000-0000-000000000001',
		14.2,
		94.0,
		10.1,
		18.5,
		'{"1": 0.8, "2": 0.9, "3": 1.1, "4": 1.4, "5": 1.8, "6": 2.1, "7": 2.0, "8": 1.7, "9": 1.3, "10": 1.0, "11": 0.9, "12": 0.8}'::jsonb,
		'{"Elk": 55.0, "Moose": 30.0, "Grizzly bear": 15.0}'::jsonb,
		'[{"feature": "Canopy cover", "impact": "High attraction", "weight": 0.45}, {"feature": "Gallatin River corridor", "impact": "High attraction", "weight": 0.35}, {"feature": "Winter snow depth", "impact": "Bottleneck", "weight": 0.30}]'::jsonb
	),
	(
		'00000000-0000-0000-0001-000000000001',
		18.2,
		97.5,
		12.5,
		24.0,
		'{"1": 1.2, "2": 1.4, "3": 1.5, "4": 1.3, "5": 1.1, "6": 1.0, "7": 0.9, "8": 1.1, "9": 1.6, "10": 2.1, "11": 2.2, "12": 1.8}'::jsonb,
		'{"Elk": 55.0, "Moose": 30.0, "Grizzly bear": 15.0}'::jsonb,
		'[{"feature": "Gallatin Canyon riparian bottleneck", "impact": "Critical", "weight": 0.65}, {"feature": "Heavy commuter traffic", "impact": "High hazard", "weight": 0.40}]'::jsonb
	),
	(
		'00000000-0000-0000-0001-000000000002',
		17.6,
		97.0,
		12.0,
		23.5,
		'{"1": 1.2, "2": 1.4, "3": 1.5, "4": 1.3, "5": 1.1, "6": 1.0, "7": 0.9, "8": 1.1, "9": 1.6, "10": 2.1, "11": 2.2, "12": 1.8}'::jsonb,
		'{"Elk": 55.0, "Moose": 30.0, "Grizzly bear": 15.0}'::jsonb,
		'[{"feature": "Gallatin Canyon riparian bottleneck", "impact": "Critical", "weight": 0.62}, {"feature": "Moose willow flats", "impact": "Attraction", "weight": 0.35}]'::jsonb
	),
	(
		'00000000-0000-0000-0001-000000000003',
		16.9,
		96.5,
		11.5,
		22.0,
		'{"1": 1.2, "2": 1.4, "3": 1.5, "4": 1.3, "5": 1.1, "6": 1.0, "7": 0.9, "8": 1.1, "9": 1.6, "10": 2.1, "11": 2.2, "12": 1.8}'::jsonb,
		'{"Elk": 55.0, "Moose": 30.0, "Grizzly bear": 15.0}'::jsonb,
		'[{"feature": "Gallatin Canyon riparian bottleneck", "impact": "High", "weight": 0.58}]'::jsonb
	),
	(
		'00000000-0000-0000-0001-000000000004',
		7.5,
		82.0,
		4.5,
		11.0,
		'{"5": 1.2, "6": 1.5, "7": 1.6, "8": 1.4, "9": 1.2}'::jsonb,
		'{"Elk": 60.0, "Moose": 40.0}'::jsonb,
		'[{"feature": "Open valley transition", "impact": "Moderate", "weight": 0.25}]'::jsonb
	),

	-- Grand Teton (US-89 MP 18.0-18.3 cluster + MP 25.0 & 35.0)
	(
		'00000000-0000-0000-0002-000000000001',
		17.4,
		96.8,
		11.8,
		23.0,
		'{"4": 1.5, "5": 2.2, "6": 1.8, "9": 1.6, "10": 2.4, "11": 2.0}'::jsonb,
		'{"Mule deer": 50.0, "Pronghorn": 30.0, "Elk": 20.0}'::jsonb,
		'[{"feature": "Path of the Pronghorn corridor", "impact": "Critical", "weight": 0.68}, {"feature": "High-speed flat straightaway", "impact": "High hazard", "weight": 0.42}]'::jsonb
	),
	(
		'00000000-0000-0000-0002-000000000002',
		16.8,
		96.2,
		11.2,
		22.5,
		'{"4": 1.5, "5": 2.2, "6": 1.8, "9": 1.6, "10": 2.4, "11": 2.0}'::jsonb,
		'{"Mule deer": 50.0, "Pronghorn": 30.0, "Elk": 20.0}'::jsonb,
		'[{"feature": "Blacktail Butte migration bottleneck", "impact": "High", "weight": 0.60}]'::jsonb
	),
	(
		'00000000-0000-0000-0002-000000000003',
		15.9,
		95.5,
		10.5,
		21.0,
		'{"4": 1.5, "5": 2.2, "6": 1.8, "9": 1.6, "10": 2.4, "11": 2.0}'::jsonb,
		'{"Mule deer": 50.0, "Pronghorn": 30.0, "Elk": 20.0}'::jsonb,
		'[{"feature": "Antelope Flats junction", "impact": "High", "weight": 0.55}]'::jsonb
	),
	(
		'00000000-0000-0000-0002-000000000004',
		8.2,
		84.0,
		5.0,
		12.0,
		'{"5": 1.4, "6": 1.7, "7": 1.5, "8": 1.3}'::jsonb,
		'{"Mule deer": 60.0, "Elk": 40.0}'::jsonb,
		'[{"feature": "Snake River bench", "impact": "Moderate", "weight": 0.30}]'::jsonb
	),
	(
		'00000000-0000-0000-0002-000000000005',
		6.5,
		78.0,
		4.0,
		9.5,
		'{"5": 1.2, "6": 1.5, "9": 1.4, "10": 1.8}'::jsonb,
		'{"Elk": 70.0, "Mule deer": 30.0}'::jsonb,
		'[{"feature": "Moran Junction approach", "impact": "Moderate", "weight": 0.28}]'::jsonb
	),

	-- Glacier (US-2 MP 152.0-152.3 cluster + MP 165.0 & 180.0)
	(
		'00000000-0000-0000-0003-000000000001',
		19.1,
		98.0,
		13.0,
		26.0,
		'{"5": 1.6, "6": 2.4, "7": 2.5, "8": 2.2, "9": 1.8, "10": 1.2}'::jsonb,
		'{"Grizzly bear": 40.0, "Canada lynx": 25.0, "Bighorn sheep": 35.0}'::jsonb,
		'[{"feature": "Goat Lick mineral attraction", "impact": "Critical", "weight": 0.72}, {"feature": "Middle Fork Flathead river bottleneck", "impact": "High attraction", "weight": 0.45}]'::jsonb
	),
	(
		'00000000-0000-0000-0003-000000000002',
		18.5,
		97.6,
		12.5,
		25.0,
		'{"5": 1.6, "6": 2.4, "7": 2.5, "8": 2.2, "9": 1.8, "10": 1.2}'::jsonb,
		'{"Grizzly bear": 40.0, "Canada lynx": 25.0, "Bighorn sheep": 35.0}'::jsonb,
		'[{"feature": "Goat Lick mineral attraction", "impact": "Critical", "weight": 0.68}, {"feature": "Steep talus crossing", "impact": "Funnel", "weight": 0.38}]'::jsonb
	),
	(
		'00000000-0000-0000-0003-000000000003',
		17.2,
		96.5,
		11.5,
		23.5,
		'{"5": 1.6, "6": 2.4, "7": 2.5, "8": 2.2, "9": 1.8, "10": 1.2}'::jsonb,
		'{"Grizzly bear": 40.0, "Canada lynx": 25.0, "Bighorn sheep": 35.0}'::jsonb,
		'[{"feature": "Essex / Pinnacle crossing", "impact": "High", "weight": 0.58}]'::jsonb
	),
	(
		'00000000-0000-0000-0003-000000000004',
		9.4,
		86.0,
		6.0,
		13.5,
		'{"6": 1.8, "7": 2.0, "8": 1.9, "9": 1.5}'::jsonb,
		'{"Grizzly bear": 50.0, "Canada lynx": 50.0}'::jsonb,
		'[{"feature": "Marias Pass Continental Divide summit", "impact": "Moderate corridor", "weight": 0.42}]'::jsonb
	),
	(
		'00000000-0000-0000-0003-000000000005',
		5.8,
		75.0,
		3.5,
		8.5,
		'{"5": 1.2, "6": 1.4, "7": 1.5, "8": 1.3}'::jsonb,
		'{"Bighorn sheep": 60.0, "Grizzly bear": 40.0}'::jsonb,
		'[{"feature": "East Glacier approach", "impact": "Low hazard", "weight": 0.20}]'::jsonb
	),

	-- Rocky Mountain (US-34 MP 58.0-58.3 cluster + MP 65.0 & 69.5)
	(
		'00000000-0000-0000-0004-000000000001',
		18.0,
		97.2,
		12.0,
		24.5,
		'{"5": 1.2, "6": 2.2, "7": 2.8, "8": 2.6, "9": 2.0, "10": 1.4}'::jsonb,
		'{"Elk": 55.0, "Mule deer": 30.0, "Bighorn sheep": 15.0}'::jsonb,
		'[{"feature": "Milner Pass Continental Divide crossing", "impact": "Critical", "weight": 0.70}, {"feature": "High summer visitor volume", "impact": "High hazard", "weight": 0.45}]'::jsonb
	),
	(
		'00000000-0000-0000-0004-000000000002',
		17.5,
		96.9,
		11.5,
		23.8,
		'{"5": 1.2, "6": 2.2, "7": 2.8, "8": 2.6, "9": 2.0, "10": 1.4}'::jsonb,
		'{"Elk": 55.0, "Mule deer": 30.0, "Bighorn sheep": 15.0}'::jsonb,
		'[{"feature": "Alpine tundra summer grazing funnel", "impact": "Critical", "weight": 0.65}]'::jsonb
	),
	(
		'00000000-0000-0000-0004-000000000003',
		16.4,
		96.0,
		11.0,
		22.5,
		'{"5": 1.2, "6": 2.2, "7": 2.8, "8": 2.6, "9": 2.0, "10": 1.4}'::jsonb,
		'{"Elk": 55.0, "Mule deer": 30.0, "Bighorn sheep": 15.0}'::jsonb,
		'[{"feature": "Poudre Lake headwaters", "impact": "High attraction", "weight": 0.52}]'::jsonb
	),
	(
		'00000000-0000-0000-0004-000000000004',
		8.8,
		85.0,
		5.5,
		12.5,
		'{"6": 2.0, "7": 2.6, "8": 2.4, "9": 1.6}'::jsonb,
		'{"Bighorn sheep": 50.0, "Elk": 50.0}'::jsonb,
		'[{"feature": "Lava Cliffs / Alpine Center", "impact": "Moderate", "weight": 0.35}]'::jsonb
	),
	(
		'00000000-0000-0000-0004-000000000005',
		6.2,
		77.0,
		3.8,
		9.0,
		'{"5": 1.2, "6": 1.5, "9": 1.8, "10": 2.0}'::jsonb,
		'{"Mule deer": 60.0, "Elk": 40.0}'::jsonb,
		'[{"feature": "Many Parks Curve", "impact": "Low hazard", "weight": 0.22}]'::jsonb
	),

	-- Yosemite (CA-120 MP 42.0-42.3 cluster + MP 65.0 & 74.9, plus CA-140)
	(
		'00000000-0000-0000-0000-000000000005',
		16.8,
		95.0,
		11.0,
		22.5,
		'{"1": 0.2, "2": 0.2, "3": 0.4, "4": 0.7, "5": 1.4, "6": 2.4, "7": 2.8, "8": 2.6, "9": 1.9, "10": 1.0, "11": 0.4, "12": 0.2}'::jsonb,
		'{"Black bear": 45.0, "Mule deer": 40.0, "Sierra bighorn sheep": 15.0}'::jsonb,
		'[{"feature": "High summer visitor volume", "impact": "Critical", "weight": 0.55}, {"feature": "Winter snow pass closure", "impact": "Seasonal relief", "weight": -0.35}]'::jsonb
	),
	(
		'00000000-0000-0000-0005-000000000002',
		17.5,
		96.8,
		11.5,
		23.5,
		'{"1": 0.2, "2": 0.2, "3": 0.4, "4": 0.7, "5": 1.4, "6": 2.4, "7": 2.8, "8": 2.6, "9": 1.9, "10": 1.0, "11": 0.4, "12": 0.2}'::jsonb,
		'{"Black bear": 45.0, "Mule deer": 40.0, "Sierra bighorn sheep": 15.0}'::jsonb,
		'[{"feature": "Crane Flat bear feeding zone", "impact": "Critical", "weight": 0.65}, {"feature": "Oak mast attraction", "impact": "High", "weight": 0.42}]'::jsonb
	),
	(
		'00000000-0000-0000-0005-000000000003',
		16.2,
		95.8,
		10.8,
		21.8,
		'{"1": 0.2, "2": 0.2, "3": 0.4, "4": 0.7, "5": 1.4, "6": 2.4, "7": 2.8, "8": 2.6, "9": 1.9, "10": 1.0, "11": 0.4, "12": 0.2}'::jsonb,
		'{"Black bear": 45.0, "Mule deer": 40.0, "Sierra bighorn sheep": 15.0}'::jsonb,
		'[{"feature": "Crane Flat meadow transition", "impact": "High", "weight": 0.52}]'::jsonb
	),
	(
		'00000000-0000-0000-0005-000000000004',
		9.1,
		85.5,
		5.8,
		13.0,
		'{"6": 2.2, "7": 2.9, "8": 2.7, "9": 1.8}'::jsonb,
		'{"Mule deer": 70.0, "Black bear": 30.0}'::jsonb,
		'[{"feature": "Tuolumne Meadows subalpine", "impact": "Moderate", "weight": 0.35}]'::jsonb
	),
	(
		'00000000-0000-0000-0005-000000000005',
		8.4,
		84.2,
		5.2,
		12.2,
		'{"6": 1.8, "7": 2.5, "8": 2.4, "9": 1.6}'::jsonb,
		'{"Sierra bighorn sheep": 60.0, "Mule deer": 40.0}'::jsonb,
		'[{"feature": "Tioga Pass summit high alpine", "impact": "Endangered species corridor", "weight": 0.48}]'::jsonb
	),
	(
		'00000000-0000-0000-0000-000000000006',
		13.5,
		92.0,
		9.0,
		17.5,
		'{"1": 0.8, "2": 0.9, "3": 1.2, "4": 1.5, "5": 1.8, "6": 1.9, "7": 1.8, "8": 1.6, "9": 1.4, "10": 1.1, "11": 0.9, "12": 0.8}'::jsonb,
		'{"Mule deer": 65.0, "Black bear": 35.0}'::jsonb,
		'[{"feature": "Merced River riparian corridor", "impact": "High attraction", "weight": 0.48}, {"feature": "Steep canyon topography", "impact": "Bottleneck", "weight": 0.32}]'::jsonb
	),

	-- Additional inter-state benchmark corridors (NV, UT, ID)
	(
		'00000000-0000-0000-0000-000000000002',
		18.5,
		96.5,
		12.0,
		24.0,
		'{"1": 1.2, "2": 1.5, "11": 2.0, "12": 1.8}'::jsonb,
		'{"Mule deer": 90.0, "Pronghorn": 10.0}'::jsonb,
		'[{"feature": "Migration corridor overlap", "impact": "Critical", "weight": 0.65}]'::jsonb
	),
	(
		'00000000-0000-0000-0000-000000000003',
		12.1,
		91.0,
		8.5,
		15.2,
		'{"5": 1.8, "6": 2.2, "7": 2.0}'::jsonb,
		'{"Pronghorn": 70.0, "Mule deer": 30.0}'::jsonb,
		'[{"feature": "Open terrain bottleneck", "impact": "High", "weight": 0.50}]'::jsonb
	),
	(
		'00000000-0000-0000-0000-000000000004',
		9.8,
		87.0,
		6.0,
		13.0,
		'{"10": 1.5, "11": 1.9, "12": 1.4}'::jsonb,
		'{"Mule deer": 60.0, "Elk": 40.0}'::jsonb,
		'[{"feature": "Canyon slope", "impact": "Medium", "weight": 0.35}]'::jsonb
	)
	ON CONFLICT (segment_id) DO UPDATE SET
		risk_score = EXCLUDED.risk_score,
		percentile_rank = EXCLUDED.percentile_rank,
		p10 = EXCLUDED.p10,
		p90 = EXCLUDED.p90,
		monthly_profile = EXCLUDED.monthly_profile,
		species_breakdown = EXCLUDED.species_breakdown,
		shap_drivers = EXCLUDED.shap_drivers,
		updated_at = now();
	`

	if _, err := tx.Exec(ctx, seedSQL); err != nil {
		return fmt.Errorf("execute seed sql: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}

	if stdout != nil {
		_, _ = fmt.Fprintln(stdout, "Seeded 28 benchmark road segments across 5 national park corridors with risk scores and SHAP drivers.")
	}
	return nil
}

func seedSegmentsCmd(ctx context.Context, stdout io.Writer) error {
	dsn := os.Getenv("CORRIDOR_MIGRATION_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("CORRIDOR_DATABASE_URL")
	}
	if dsn == "" {
		return errors.New("CORRIDOR_MIGRATION_DATABASE_URL or CORRIDOR_DATABASE_URL required")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	return SeedSegments(ctx, pool, stdout)
}

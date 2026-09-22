package predict

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Samples(context.Context) ([]Sample, error)
}
type Store struct{ Pool *pgxpool.Pool }

func (s Store) Samples(ctx context.Context) ([]Sample, error) {
	rows, err := s.Pool.Query(ctx, `SELECT cell,block,x,y,routes,road,crossings,ST_AsGeoJSON(geom)::json FROM public_movement_features ORDER BY cell LIMIT $1`, MaxCells+1)
	if err != nil {
		return nil, fmt.Errorf("read model support: %w", err)
	}
	defer rows.Close()
	result := []Sample{}
	for rows.Next() {
		var v Sample
		if err = rows.Scan(&v.Cell, &v.Block, &v.X, &v.Y, &v.Routes, &v.Road, &v.Crossings, &v.Geometry); err != nil {
			return nil, fmt.Errorf("decode model support: %w", err)
		}
		result = append(result, v)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("complete model support: %w", err)
	}
	return result, nil
}

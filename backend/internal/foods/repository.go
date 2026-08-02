package foods

import (
	"context"

	"github.com/dhruvv301292/nutrichat/internal/nutrition"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context) ([]nutrition.Food, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT name, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit
		FROM foods
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = []nutrition.Food{}
	for rows.Next() {
		var f nutrition.Food
		if err := rows.Scan(&f.Name, &f.Calories, &f.Protein, &f.Carbs, &f.Fat, &f.Fiber, &f.Sodium, &f.Unit, &f.UnitQuantity, &f.GramsPerUnit); err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) Search(ctx context.Context, query string) ([]nutrition.Food, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT name, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit
		FROM foods
		WHERE similarity(name, $1) > 0.2
		ORDER BY similarity(name, $1) DESC
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = []nutrition.Food{}
	for rows.Next() {
		var f nutrition.Food
		if err := rows.Scan(&f.Name, &f.Calories, &f.Protein, &f.Carbs, &f.Fat, &f.Fiber, &f.Sodium, &f.Unit, &f.UnitQuantity, &f.GramsPerUnit); err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

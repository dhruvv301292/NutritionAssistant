package foods

import (
	"context"
	"errors"

	"github.com/dhruvv301292/nutrichat/internal/nutrition"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("food not found")

// similarityThreshold is the minimum pg_trgm score a food name needs to
// count as a candidate match. 0.2 (the previous value) was too permissive:
// generic multi-word queries like "chicken wings" would trigram-match an
// unrelated food like "Chicken Thigh, Raw" (0.33) just from sharing the
// word "chicken", which silently returned the wrong food AND prevented the
// external-database fallback from ever running for a food we don't have.
const similarityThreshold = 0.4

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context) ([]nutrition.Food, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit
		FROM foods
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = []nutrition.Food{}
	for rows.Next() {
		var f nutrition.Food
		if err := rows.Scan(&f.ID, &f.Name, &f.Calories, &f.Protein, &f.Carbs, &f.Fat, &f.Fiber, &f.Sodium, &f.Unit, &f.UnitQuantity, &f.GramsPerUnit); err != nil {
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
		SELECT id, name, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit
		FROM foods
		WHERE similarity(name, $1) > $2
		ORDER BY similarity(name, $1) DESC
	`, query, similarityThreshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = []nutrition.Food{}
	for rows.Next() {
		var f nutrition.Food
		if err := rows.Scan(&f.ID, &f.Name, &f.Calories, &f.Protein, &f.Carbs, &f.Fat, &f.Fiber, &f.Sodium, &f.Unit, &f.UnitQuantity, &f.GramsPerUnit); err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

type ScoredFood struct {
	Food       nutrition.Food
	Similarity float64
}

func (r *Repository) SearchWithScore(ctx context.Context, query string) ([]ScoredFood, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit, similarity(name, $1)
		FROM foods
		WHERE similarity(name, $1) > $2
		ORDER BY similarity(name, $1) DESC
	`, query, similarityThreshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = []ScoredFood{}
	for rows.Next() {
		var sf ScoredFood
		if err := rows.Scan(&sf.Food.ID, &sf.Food.Name, &sf.Food.Calories, &sf.Food.Protein, &sf.Food.Carbs, &sf.Food.Fat, &sf.Food.Fiber, &sf.Food.Sodium, &sf.Food.Unit, &sf.Food.UnitQuantity, &sf.Food.GramsPerUnit, &sf.Similarity); err != nil {
			return nil, err
		}
		result = append(result, sf)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// Insert adds a new food row, used by the cache-aside fallback when an
// external food database resolves something Postgres didn't have yet (see
// CLAUDE.md's Food Lookup Flow). ON CONFLICT handles the race where two
// concurrent requests both miss the DB and both fetch the same food
// externally — the second insert is dropped and its row re-fetched by name,
// rather than erroring or duplicating the food.
func (r *Repository) Insert(ctx context.Context, f nutrition.Food) (nutrition.Food, error) {
	var inserted nutrition.Food
	err := r.pool.QueryRow(ctx, `
		INSERT INTO foods (name, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit
	`, f.Name, f.Calories, f.Protein, f.Carbs, f.Fat, f.Fiber, f.Sodium, f.Unit, f.UnitQuantity, f.GramsPerUnit).
		Scan(&inserted.ID, &inserted.Name, &inserted.Calories, &inserted.Protein, &inserted.Carbs, &inserted.Fat, &inserted.Fiber, &inserted.Sodium, &inserted.Unit, &inserted.UnitQuantity, &inserted.GramsPerUnit)
	if err != nil {
		return nutrition.Food{}, err
	}
	return inserted, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (nutrition.Food, error) {
	var f nutrition.Food
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit
		FROM foods
		WHERE id = $1
	`, id).Scan(&f.ID, &f.Name, &f.Calories, &f.Protein, &f.Carbs, &f.Fat, &f.Fiber, &f.Sodium, &f.Unit, &f.UnitQuantity, &f.GramsPerUnit)
	if errors.Is(err, pgx.ErrNoRows) {
		return nutrition.Food{}, ErrNotFound
	}
	if err != nil {
		return nutrition.Food{}, err
	}
	return f, nil
}

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
// count as a candidate match. 0.2 (an earlier value) was too permissive:
// generic multi-word queries like "chicken wings" would trigram-match an
// unrelated food like "Chicken Thigh, Raw" (0.33) just from sharing the
// word "chicken", which silently returned the wrong food AND prevented the
// external-database fallback from ever running for a food we don't have.
const similarityThreshold = 0.4

// brandHardThreshold is the minimum brand-name trigram similarity required
// to accept a candidate when the query specifies a brand. This is a hard
// filter, separate from and stricter than the overall name similarity:
// "protein shake" alone is expected to fuzzy-match across many brands, but
// once a brand is specified, a different brand must never pass ("optimum
// nutrition" vs "quest" scores far below this). See searchQuery/matcher.go.
const brandHardThreshold = 0.6

// brandAmbiguousLow/High bound the zone where brand similarity is neither a
// confident match nor a confident non-match — e.g. an abbreviation or minor
// misspelling. Candidates whose brand score falls in this band are resolved
// by an LLM verifier (see Matcher.resolveItem) instead of being silently
// accepted or rejected by the trigram score alone.
const brandAmbiguousLow = 0.4
const brandAmbiguousHigh = brandHardThreshold

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context) ([]nutrition.Food, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, brand, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit
		FROM foods
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = []nutrition.Food{}
	for rows.Next() {
		var f nutrition.Food
		if err := rows.Scan(&f.ID, &f.Name, &f.Brand, &f.Calories, &f.Protein, &f.Carbs, &f.Fat, &f.Fiber, &f.Sodium, &f.Unit, &f.UnitQuantity, &f.GramsPerUnit); err != nil {
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
		SELECT id, name, brand, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit
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
		if err := rows.Scan(&f.ID, &f.Name, &f.Brand, &f.Calories, &f.Protein, &f.Carbs, &f.Fat, &f.Fiber, &f.Sodium, &f.Unit, &f.UnitQuantity, &f.GramsPerUnit); err != nil {
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
	// BrandSimilarity is only meaningful when the query specified a brand
	// (see SearchBranded); zero otherwise.
	BrandSimilarity float64
}

// SearchWithScore is the brand-agnostic candidate search: pure trigram
// similarity on food name only. Used when the query has no brand context.
func (r *Repository) SearchWithScore(ctx context.Context, query string) ([]ScoredFood, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, brand, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit, similarity(name, $1)
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
		if err := rows.Scan(&sf.Food.ID, &sf.Food.Name, &sf.Food.Brand, &sf.Food.Calories, &sf.Food.Protein, &sf.Food.Carbs, &sf.Food.Fat, &sf.Food.Fiber, &sf.Food.Sodium, &sf.Food.Unit, &sf.Food.UnitQuantity, &sf.Food.GramsPerUnit, &sf.Similarity); err != nil {
			return nil, err
		}
		result = append(result, sf)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// SearchBranded is the two-stage retrieval used when the query specifies a
// brand: candidate generation via combined trigram similarity over
// "brand name" (fast, typo-forgiving), paired with the brand-only
// similarity scored separately so the caller can apply it as a hard filter.
// Candidates are NOT filtered by brandHardThreshold here — Postgres ranks
// and returns the top candidates, but accept/reject/verify decisions belong
// to the caller (Matcher), since the ambiguous-zone case needs an LLM
// verification step this repository has no business making.
func (r *Repository) SearchBranded(ctx context.Context, brand, productName string) ([]ScoredFood, error) {
	combined := brand + " " + productName
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, brand, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit,
		       similarity(coalesce(brand, '') || ' ' || name, $1) AS combined_sim,
		       similarity(coalesce(brand, ''), $2) AS brand_sim
		FROM foods
		WHERE similarity(coalesce(brand, '') || ' ' || name, $1) > $3
		ORDER BY combined_sim DESC
		LIMIT 10
	`, combined, brand, similarityThreshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = []ScoredFood{}
	for rows.Next() {
		var sf ScoredFood
		if err := rows.Scan(&sf.Food.ID, &sf.Food.Name, &sf.Food.Brand, &sf.Food.Calories, &sf.Food.Protein, &sf.Food.Carbs, &sf.Food.Fat, &sf.Food.Fiber, &sf.Food.Sodium, &sf.Food.Unit, &sf.Food.UnitQuantity, &sf.Food.GramsPerUnit, &sf.Similarity, &sf.BrandSimilarity); err != nil {
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
// externally — the second insert is dropped and its row re-fetched by
// brand+name, rather than erroring or duplicating the food. Note: this only
// dedupes when brand is non-NULL — Postgres unique constraints never match
// NULL against NULL, so concurrent inserts of two different unbranded foods
// that happen to share a name can still both succeed. Acceptable: unbranded
// name collisions are rare and harmless (worst case, a duplicate row).
func (r *Repository) Insert(ctx context.Context, f nutrition.Food) (nutrition.Food, error) {
	var inserted nutrition.Food
	err := r.pool.QueryRow(ctx, `
		INSERT INTO foods (name, brand, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (brand, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, brand, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit
	`, f.Name, f.Brand, f.Calories, f.Protein, f.Carbs, f.Fat, f.Fiber, f.Sodium, f.Unit, f.UnitQuantity, f.GramsPerUnit).
		Scan(&inserted.ID, &inserted.Name, &inserted.Brand, &inserted.Calories, &inserted.Protein, &inserted.Carbs, &inserted.Fat, &inserted.Fiber, &inserted.Sodium, &inserted.Unit, &inserted.UnitQuantity, &inserted.GramsPerUnit)
	if err != nil {
		return nutrition.Food{}, err
	}
	return inserted, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (nutrition.Food, error) {
	var f nutrition.Food
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, brand, calories, protein, carbs, fat, fiber, sodium, unit, unit_quantity, grams_per_unit
		FROM foods
		WHERE id = $1
	`, id).Scan(&f.ID, &f.Name, &f.Brand, &f.Calories, &f.Protein, &f.Carbs, &f.Fat, &f.Fiber, &f.Sodium, &f.Unit, &f.UnitQuantity, &f.GramsPerUnit)
	if errors.Is(err, pgx.ErrNoRows) {
		return nutrition.Food{}, ErrNotFound
	}
	if err != nil {
		return nutrition.Food{}, err
	}
	return f, nil
}

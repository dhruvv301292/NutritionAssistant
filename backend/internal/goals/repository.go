package goals

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Get returns the user's stored goals, or sensible defaults if they've
// never set any — callers don't need to special-case a "no goals yet" state.
func (r *Repository) Get(ctx context.Context, userID int) (Goals, error) {
	var g Goals
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, calorie_goal, protein_goal, carb_goal, fat_goal, fiber_goal, tracked_macros
		FROM user_goals WHERE user_id = $1
	`, userID).Scan(&g.UserID, &g.CalorieGoal, &g.ProteinGoal, &g.CarbGoal, &g.FatGoal, &g.FiberGoal, &g.TrackedMacros)
	if errors.Is(err, pgx.ErrNoRows) {
		return Defaults(userID), nil
	}
	if err != nil {
		return Goals{}, err
	}
	return g, nil
}

func (r *Repository) Upsert(ctx context.Context, g Goals) (Goals, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_goals (user_id, calorie_goal, protein_goal, carb_goal, fat_goal, fiber_goal, tracked_macros, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
		ON CONFLICT (user_id) DO UPDATE SET
			calorie_goal = EXCLUDED.calorie_goal,
			protein_goal = EXCLUDED.protein_goal,
			carb_goal = EXCLUDED.carb_goal,
			fat_goal = EXCLUDED.fat_goal,
			fiber_goal = EXCLUDED.fiber_goal,
			tracked_macros = EXCLUDED.tracked_macros,
			updated_at = now()
	`, g.UserID, g.CalorieGoal, g.ProteinGoal, g.CarbGoal, g.FatGoal, g.FiberGoal, g.TrackedMacros)
	if err != nil {
		return Goals{}, err
	}
	return g, nil
}

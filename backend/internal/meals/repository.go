package meals

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Save inserts a meal log and its items in a single transaction so a
// failure partway through doesn't leave a meal_logs row with no items.
func (r *Repository) Save(ctx context.Context, userID int, items []LogItem) (Log, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Log{}, err
	}
	defer tx.Rollback(ctx)

	log := Log{UserID: userID}
	err = tx.QueryRow(ctx, `
		INSERT INTO meal_logs (user_id) VALUES ($1)
		RETURNING id, user_id, logged_at
	`, userID).Scan(&log.ID, &log.UserID, &log.LoggedAt)
	if err != nil {
		return Log{}, err
	}

	for _, item := range items {
		var savedItem LogItem
		err = tx.QueryRow(ctx, `
			INSERT INTO meal_log_items (meal_log_id, food_id, quantity, unit)
			VALUES ($1, $2, $3, $4)
			RETURNING id, food_id, quantity, unit
		`, log.ID, item.FoodID, item.Quantity, item.Unit).Scan(&savedItem.ID, &savedItem.FoodID, &savedItem.Quantity, &savedItem.Unit)
		if err != nil {
			return Log{}, err
		}
		log.Items = append(log.Items, savedItem)
	}

	if err := tx.Commit(ctx); err != nil {
		return Log{}, err
	}
	return log, nil
}

// ListByUserAndDateRange fetches meal logs (with items) for a user whose
// logged_at falls within [start, end).
func (r *Repository) ListByUserAndDateRange(ctx context.Context, userID int, start, end string) ([]Log, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, logged_at
		FROM meal_logs
		WHERE user_id = $1 AND logged_at >= $2 AND logged_at < $3
		ORDER BY logged_at
	`, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []Log{}
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.ID, &l.UserID, &l.LoggedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range logs {
		items, err := r.itemsForLog(ctx, logs[i].ID)
		if err != nil {
			return nil, err
		}
		logs[i].Items = items
	}

	return logs, nil
}

func (r *Repository) itemsForLog(ctx context.Context, mealLogID int) ([]LogItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, food_id, quantity, unit
		FROM meal_log_items
		WHERE meal_log_id = $1
	`, mealLogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []LogItem{}
	for rows.Next() {
		var item LogItem
		if err := rows.Scan(&item.ID, &item.FoodID, &item.Quantity, &item.Unit); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

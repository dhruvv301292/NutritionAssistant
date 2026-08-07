package users

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

// UpsertGoogleUser finds a user by google_sub, or by email if this is the
// first time they've signed in with Google on an account that previously
// only existed via seed data, and otherwise creates a new one.
func (r *Repository) UpsertGoogleUser(ctx context.Context, googleSub, email, name string) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (google_sub, email, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE
			SET google_sub = EXCLUDED.google_sub,
				name = EXCLUDED.name
		RETURNING id, google_sub, email, name, created_at
	`, googleSub, email, name).Scan(&u.ID, &u.GoogleSub, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *Repository) FindByID(ctx context.Context, id int) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, google_sub, email, name, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(&u.ID, &u.GoogleSub, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

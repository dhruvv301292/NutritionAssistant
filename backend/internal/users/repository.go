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
		RETURNING id, google_sub, apple_sub, email, COALESCE(name, ''), created_at
	`, googleSub, email, name).Scan(&u.ID, &u.GoogleSub, &u.AppleSub, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

// UpsertAppleUser finds a user by apple_sub, or by email if this is the
// first time they've signed in with Apple on an account that previously
// only existed via seed data or Google sign-in, and otherwise creates a new
// one. name is only non-empty on a user's very first-ever Apple sign-in —
// Apple never resends it — so an empty name here must never overwrite an
// existing one (COALESCE keeps whatever's already on file).
func (r *Repository) UpsertAppleUser(ctx context.Context, appleSub, email, name string) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (apple_sub, email, name)
		VALUES ($1, $2, NULLIF($3, ''))
		ON CONFLICT (email) DO UPDATE
			SET apple_sub = EXCLUDED.apple_sub,
				name = COALESCE(NULLIF(EXCLUDED.name, ''), users.name)
		RETURNING id, google_sub, apple_sub, email, COALESCE(name, ''), created_at
	`, appleSub, email, name).Scan(&u.ID, &u.GoogleSub, &u.AppleSub, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *Repository) FindByID(ctx context.Context, id int) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `
		SELECT id, google_sub, apple_sub, email, COALESCE(name, ''), created_at
		FROM users
		WHERE id = $1
	`, id).Scan(&u.ID, &u.GoogleSub, &u.AppleSub, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

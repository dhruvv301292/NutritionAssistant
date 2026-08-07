package users

import "time"

type User struct {
	ID        int       `json:"id"`
	GoogleSub *string   `json:"-"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

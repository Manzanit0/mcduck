package users

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/manzanit0/mcduck/pkg/auth"
)

type User struct {
	Email          string `db:"email"`
	HashedPassword string `db:"hashed_password"`
	Password       string
}

func Create(ctx context.Context, db *sqlx.DB, u User) (User, error) {
	hashed, err := auth.HashPassword(u.Password)
	if err != nil {
		return u, fmt.Errorf("could not hash password: %w", err)
	}

	u.HashedPassword = hashed

	_, err = db.ExecContext(ctx, `INSERT INTO users (email, hashed_password) VALUES ($1, $2)`, u.Email, u.HashedPassword)
	if err != nil {
		return u, err
	}

	return u, nil
}

func Find(ctx context.Context, db *sqlx.DB, email string) (*User, error) {
	var u User
	err := db.GetContext(ctx, &u, `SELECT email, hashed_password FROM users WHERE email = $1`, email)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

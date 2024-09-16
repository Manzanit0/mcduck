package users

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/xtrace"
)

type User struct {
	Email          string `db:"email"`
	HashedPassword string `db:"hashed_password"`
	Password       string
	TelegramChatID *int64 `db:"telegram_chat_id"`
}

func Create(ctx context.Context, db *sqlx.DB, u User) (User, error) {
	ctx, span := xtrace.StartSpan(ctx, "Create User")
	defer span.End()

	hashed, err := auth.HashPassword(u.Password)
	if err != nil {
		return u, fmt.Errorf("could not hash password: %w", err)
	}

	u.HashedPassword = hashed

	_, err = db.ExecContext(ctx, `INSERT INTO users (email, hashed_password, telegram_chat_id) VALUES ($1, $2, $3)`, u.Email, u.HashedPassword, u.TelegramChatID)
	if err != nil {
		return u, err
	}

	return u, nil
}

func Find(ctx context.Context, db *sqlx.DB, email string) (*User, error) {
	ctx, span := xtrace.StartSpan(ctx, "Find User")
	defer span.End()

	var u User
	err := db.GetContext(ctx, &u, `SELECT email, hashed_password, telegram_chat_id FROM users WHERE email = $1`, email)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func FindByChatID(ctx context.Context, db *sqlx.DB, chatID int64) (*User, error) {
	ctx, span := xtrace.StartSpan(ctx, "Find User By Telegram ID")
	defer span.End()

	var u User
	err := db.GetContext(ctx, &u, `SELECT email, hashed_password, telegram_chat_id FROM users WHERE telegram_chat_id = $1`, chatID)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func UpdateTelegramChatID(ctx context.Context, db *sqlx.DB, u *User, chatID int64) error {
	ctx, span := xtrace.StartSpan(ctx, "Update User Telegram ID")
	defer span.End()

	_, err := db.ExecContext(ctx, `UPDATE users SET telegram_chat_id=$1 WHERE email=$2`, chatID, u.Email)
	if err != nil {
		return err
	}

	u.TelegramChatID = &chatID
	return nil
}

package servers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/jmoiron/sqlx"
	usersv1 "github.com/manzanit0/mcduck/api/users.v1"
	"github.com/manzanit0/mcduck/api/users.v1/usersv1connect"
	"github.com/manzanit0/mcduck/internal/users"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type usersServer struct {
	db *sqlx.DB
}

var _ usersv1connect.UsersServiceClient = &usersServer{}

func NewUsersServer(db *sqlx.DB) usersv1connect.UsersServiceClient {
	return &usersServer{db: db}
}

func (s *usersServer) GetUser(ctx context.Context, req *connect.Request[usersv1.GetUserRequest]) (*connect.Response[usersv1.GetUserResponse], error) {
	span := trace.SpanFromContext(ctx)

	u, err := users.FindByChatID(ctx, s.db, req.Msg.TelegramChatId)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		span.SetStatus(codes.Error, err.Error())
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user with telegram ID %d doesn't exist", req.Msg.TelegramChatId))
	} else if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to get user", "error", err.Error())
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to get receipt: %w", err))
	}

	res := connect.NewResponse(&usersv1.GetUserResponse{
		User: &usersv1.User{
			Email:          u.Email,
			TelegramChatId: *u.TelegramChatID,
			HashedPassword: u.HashedPassword,
		},
	})
	return res, nil
}

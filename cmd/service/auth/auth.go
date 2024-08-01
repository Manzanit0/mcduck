package auth

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/jmoiron/sqlx"
	authv1 "github.com/manzanit0/mcduck/api/auth.v1"
	"github.com/manzanit0/mcduck/api/auth.v1/authv1connect"
	"github.com/manzanit0/mcduck/internal/users"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/tgram"
)

type Server struct {
	DB       *sqlx.DB
	Telegram tgram.Client
}

var _ authv1connect.AuthServiceClient = (*Server)(nil)

func (s *Server) Register(ctx context.Context, req *connect.Request[authv1.RegisterRequest]) (*connect.Response[authv1.RegisterResponse], error) {
	user, err := users.Create(ctx, s.DB, users.User{Email: req.Msg.Email, Password: req.Msg.Password})
	if err != nil {
		slog.ErrorContext(ctx, "create user", "error", err.Error())
		return nil, fmt.Errorf("unable to create user: %w", err)
	}

	token, err := auth.GenerateJWT(user.Email)
	if err != nil {
		slog.ErrorContext(ctx, "generate JWT", "error", err.Error())
		return nil, fmt.Errorf("unable to generate token: %w", err)
	}

	res := connect.NewResponse(&authv1.RegisterResponse{
		Token: token,
	})
	return res, nil
}

func (s *Server) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	slog.Info("Got login!!")
	user, err := users.Find(ctx, s.DB, req.Msg.Email)
	if err != nil {
		slog.Error("unable to find user", "email", req.Msg.Email, "error", err.Error())
		return nil, fmt.Errorf("invalid email or password")
	}

	if !auth.CheckPasswordHash(req.Msg.Password, user.HashedPassword) {
		slog.Error("invalid password", "email", req.Msg.Email, "error", "hashed password doesn't match")
		return nil, fmt.Errorf("invalid email or password")
	}

	token, err := auth.GenerateJWT(user.Email)
	if err != nil {
		slog.ErrorContext(ctx, "generate JWT", "error", err.Error())
		return nil, fmt.Errorf("unable to generate token: %w", err)
	}

	res := connect.NewResponse(&authv1.LoginResponse{
		Token: token,
	})

	return res, nil
}

func (s *Server) ConnectTelegram(ctx context.Context, req *connect.Request[authv1.ConnectTelegramRequest]) (*connect.Response[authv1.ConnectTelegramResponse], error) {
	user, err := users.Find(ctx, s.DB, req.Msg.Email)
	if err != nil {
		slog.ErrorContext(ctx, "find user", "error", err.Error())
		return nil, fmt.Errorf("find user: %w", err)
	}

	err = users.UpdateTelegramChatID(ctx, s.DB, user, req.Msg.ChatId)
	if err != nil {
		slog.ErrorContext(ctx, "update telegram chat ID", "error", err.Error())
		return nil, fmt.Errorf("update user record: %w", err)
	}

	err = s.Telegram.SendMessage(tgram.SendMessageRequest{
		ChatID:    req.Msg.ChatId,
		Text:      "You account has been successfully linked\\!",
		ParseMode: tgram.ParseModeMarkdownV2,
	})
	if err != nil {
		slog.ErrorContext(ctx, "send telegram message", "error", err.Error())
		return nil, fmt.Errorf("Your account has been linked successfully but we were unable to notify you via Telegram: %w", err)
	}

	res := connect.NewResponse(&authv1.ConnectTelegramResponse{})
	return res, nil
}

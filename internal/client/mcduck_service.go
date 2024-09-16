package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/xhttp"
)

type McDuckClient interface {
	SearchUserByChatID(ctx context.Context, onBehalfOfEmail string, chatID int) (*SearchUserResponse, error)
}

type client struct {
	Host string
	h    *http.Client
}

var _ McDuckClient = (*client)(nil)

func NewMcDuckClient(host string) *client {
	h := xhttp.NewClient()
	return &client{Host: host, h: h}
}

type CreateReceiptResponse struct {
	ReceiptID int64              `json:"receipt_id"`
	Amounts   map[string]float64 `json:"receipt_amounts"`
}

type SearchUserResponse struct {
	User struct {
		Email          string `json:"email"`
		TelegramChatID int    `json:"telegram_chat_id"`
	} `json:"user"`
}

func (c *client) SearchUserByChatID(ctx context.Context, onBehalfOfEmail string, chatID int) (*SearchUserResponse, error) {
	url := fmt.Sprintf("%s/users?chat_id=%d", c.Host, chatID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	token, err := auth.GenerateJWT(onBehalfOfEmail)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s - %s", res.Status, string(respBody))
	}

	var unmarshalled SearchUserResponse
	err = json.Unmarshal(respBody, &unmarshalled)
	if err != nil {
		slog.Debug("body which was unmarsheable", "body", string(respBody), "error", err.Error())
		return nil, fmt.Errorf("unmarshal response body: %w", err)
	}

	return &unmarshalled, nil
}

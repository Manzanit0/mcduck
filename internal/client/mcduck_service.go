package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/manzanit0/mcduck/pkg/auth"
)

type McDuckClient interface {
	CreateReceipt(ctx context.Context, onBehalfOfEmail string, data []byte) (*CreateReceiptResponse, error)
	SearchUserByChatID(ctx context.Context, onBehalfOfEmail string, chatID int) (*SearchUserResponse, error)
}

type client struct {
	Host string
	h    *http.Client
}

var _ McDuckClient = (*client)(nil)

func NewMcDuckClient(host string) *client {
	return &client{Host: host, h: http.DefaultClient}
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
		log.Println("body which was unmarsheable", string(respBody))
		return nil, fmt.Errorf("unmarshal response body: %w", err)
	}

	return &unmarshalled, nil
}

func (c *client) CreateReceipt(ctx context.Context, onBehalfOfEmail string, data []byte) (*CreateReceiptResponse, error) {
	req, err := c.newCreateReceiptRequest(ctx, data, onBehalfOfEmail)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	res, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if res.StatusCode != 201 {
		return nil, fmt.Errorf("request failed: %s - %s", res.Status, string(respBody))
	}

	var unmarshalled CreateReceiptResponse
	err = json.Unmarshal(respBody, &unmarshalled)
	if err != nil {
		log.Println("body which was unmarsheable", string(respBody))
		return nil, fmt.Errorf("unmarshal response body: %w", err)
	}

	return &unmarshalled, nil
}

func (c *client) newCreateReceiptRequest(ctx context.Context, data []byte, email string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "receipt", "receipt.jpg"))
	h.Set("Content-Type", "image/jpeg")

	fw, err := writer.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	_, err = fw.Write(data)
	if err != nil {
		return nil, fmt.Errorf("copy form file to writer: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("close multipart request body writer: %w", err)
	}

	url := fmt.Sprintf("%s/receipts", c.Host)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	token, err := auth.GenerateJWT(email)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Content-Length", fmt.Sprintf("%d", body.Len()))

	return req, nil
}

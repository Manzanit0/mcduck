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
)

type McDuckClient interface {
	CreateReceipt(ctx context.Context, data []byte) (*CreateReceiptResponse, error)
}

type client struct {
	AuthToken string
	Host      string
	h         *http.Client
}

var _ McDuckClient = (*client)(nil)

func NewMcDuckClient(host, token string) *client {
	return &client{Host: host, AuthToken: token, h: http.DefaultClient}
}

type CreateReceiptResponse struct {
	ReceiptID int64              `json:"receipt_id"`
	Amounts   map[string]float64 `json:"receipt_amounts"`
}

func (c *client) CreateReceipt(ctx context.Context, data []byte) (*CreateReceiptResponse, error) {
	req, err := c.newCreateReceiptRequest(ctx, data)
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

	if res.StatusCode != 200 {
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

func (c *client) newCreateReceiptRequest(ctx context.Context, data []byte) (*http.Request, error) {
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

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Content-Length", fmt.Sprintf("%d", body.Len()))
	req.Header.Set("Authorization", c.AuthToken)

	return req, nil
}

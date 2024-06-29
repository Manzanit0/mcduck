package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/manzanit0/mcduck/pkg/auth"
)

type ParserClient interface {
	ParseReceipt(ctx context.Context, onBehalfOfEmail string, data []byte) (*ParseReceiptResponse, error)
}

type ParseReceiptResponse struct {
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Description  string  `json:"description"`
	Vendor       string  `json:"vendor"`
	PurchaseDate string  `json:"purchase_date"`
}

type parserClient struct {
	Host string
	h    *http.Client
}

var _ ParserClient = (*parserClient)(nil)

func NewParserClient(host string) *parserClient {
	return &parserClient{Host: host, h: http.DefaultClient}
}

func (c *parserClient) ParseReceipt(ctx context.Context, onBehalfOfEmail string, data []byte) (*ParseReceiptResponse, error) {
	req, err := c.newParseReceiptRequest(ctx, data, onBehalfOfEmail)
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

	var unmarshalled ParseReceiptResponse
	err = json.Unmarshal(respBody, &unmarshalled)
	if err != nil {
		slog.Debug("body which was unmarsheable", "body", string(respBody), "error", err.Error())
		return nil, fmt.Errorf("unmarshal response body: %w", err)
	}

	return &unmarshalled, nil
}

func (c *parserClient) newParseReceiptRequest(ctx context.Context, data []byte, onBehalfOf string) (*http.Request, error) {
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

	url := fmt.Sprintf("%s/receipt", c.Host)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	token, err := auth.GenerateJWT(onBehalfOf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Content-Length", fmt.Sprintf("%d", body.Len()))

	return req, nil
}

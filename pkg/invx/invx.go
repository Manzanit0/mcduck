package invx

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
)

type Client interface {
	ParseReceipt(ctx context.Context, data []byte) (map[string]float64, error)
}

type client struct {
	AuthToken string
	Host      string
	h         *http.Client
}

func NewClient(host, token string) *client {
	return &client{Host: host, AuthToken: token, h: http.DefaultClient}
}

func (c *client) ParseReceipt(_ context.Context, data []byte) (map[string]float64, error) {
	req, err := c.newParseReceiptRequest(data)
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

	var unmarshalled struct {
		Items      map[string]interface{} `json:"items"`
		ItemsCount int                    `json:"items_count"`
		Total      float64                `json:"total_price"`
	}

	err = json.Unmarshal(respBody, &unmarshalled)
	if err != nil {
		slog.Error("body which was unmarsheable", "body", string(respBody), "error", err.Error())
		return nil, fmt.Errorf("unmarshal response body: %w", err)
	}

	// The API may return strings and what not. We only care about the digits.
	retVal := make(map[string]float64)
	for k, v := range unmarshalled.Items {
		switch vv := v.(type) {
		case float64:
			retVal[k] = vv

		case float32:
			retVal[k] = float64(vv)
		}
	}

	return retVal, nil
}

func (c *client) newParseReceiptRequest(data []byte) (*http.Request, error) {
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

	url := fmt.Sprintf("%s/api/receipts/parse", c.Host)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Content-Length", fmt.Sprintf("%d", body.Len()))
	req.Header.Set("Authorization", c.AuthToken)

	return req, nil
}

package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/manzanit0/mcduck/pkg/xhttp"
	"go.opentelemetry.io/otel"
)

// -- Request structures
type Request struct {
	Model     string     `json:"model"`
	Messages  []Messages `json:"messages"`
	MaxTokens int        `json:"max_tokens"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type Content struct {
	Type     string   `json:"type"`
	Text     string   `json:"text,omitempty"`
	ImageURL ImageURL `json:"image_url,omitempty"`
}

type Messages struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// -- Response structures
type Response struct {
	ID                string    `json:"id"`
	Object            string    `json:"object"`
	Created           int       `json:"created"`
	Model             string    `json:"model"`
	Choices           []Choices `json:"choices"`
	Usage             Usage     `json:"usage"`
	SystemFingerprint string    `json:"system_fingerprint"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choices struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	Logprobs     any     `json:"logprobs"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Receipt struct {
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Description  string  `json:"description"`
	Vendor       string  `json:"vendor"`
	PurchaseDate string  `json:"purchase_date"`
}

func Completions(ctx context.Context, openaiToken string, request Request) (*Response, error) {
	tp := otel.GetTracerProvider().Tracer("parser")
	ctx, span := tp.Start(ctx, "OpenAI: Prompt Chat Completion")
	defer span.End()

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", openaiToken),
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := xhttp.NewClient()

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
}

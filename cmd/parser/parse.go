package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/manzanit0/mcduck/pkg/xhttp"
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

func parseReceiptImage(ctx context.Context, openaiToken string, imageData []byte) (*Receipt, error) {
	initialPrompt := `
You are an assistant that can read all kind of receipts and extract its
contents.

You will provide the total price paid, the currency, a summary of the items
purchased and the vendor name.

When available, you shall also provide the purchase date in the dd/MM/yyyy
format.

You will provide all this in JSON format where they property names are
"amount", "currency", "description", "purchase_date" and "vendor".

The description should not be an enumeration of the receipt contents. It should
be a summary of twenty words top of what was purchased.

The total price paid should not include the currency and it should be formated
as a number.

The currency will be formatted following the ISO 4217 codes.
`

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", openaiToken),
	}

	payload := Request{
		Model:     "gpt-4o",
		MaxTokens: 300,
		Messages: []Messages{
			{
				Role: "user",
				Content: []Content{
					{
						Type: "text",
						Text: initialPrompt,
					},
					{
						Type: "image_url",
						ImageURL: ImageURL{
							URL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
						},
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := xhttp.NewClient()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %w", err)
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("Error unmarshalling response: %w", err)
	}

	j := trimMarkdownWrapper(result.Choices[0].Message.Content)

	var receipt Receipt
	err = json.Unmarshal([]byte(j), &receipt)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling receipt: %w", err)
	}

	return &receipt, nil
}

func trimMarkdownWrapper(s string) string {
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimSuffix(s, "```")
	return s
}

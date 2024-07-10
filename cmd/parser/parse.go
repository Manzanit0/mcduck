package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/manzanit0/mcduck/pkg/xhttp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
)

const initialPrompt = `
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

func parseReceiptPDF(ctx context.Context, openaiToken string, fileName string, imageData []byte) (*Receipt, error) {
	config, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-west-1"))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	s3Svc := s3.NewFromConfig(config)

	_, err = s3Svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("scratch-go"),
		Key:    aws.String(fileName),
		Body:   bytes.NewBuffer(imageData),
	})
	if err != nil {
		return nil, fmt.Errorf("putting receipt to S3: %w", err)
	}

	svc := textract.NewFromConfig(config)
	resp, err := svc.StartDocumentTextDetection(ctx, &textract.StartDocumentTextDetectionInput{
		DocumentLocation: &types.DocumentLocation{
			S3Object: &types.S3Object{
				Bucket: aws.String("scratch-go"),
				Name:   aws.String(fileName),
			},
		},
		JobTag: aws.String("scratch-go"),
	})
	if err != nil {
		return nil, fmt.Errorf("starting text detection: %w", err)
	}

	var out *textract.GetDocumentTextDetectionOutput

	// poll AWS Textract until the job has finished.
outer:
	for {
		select {

		case <-ctx.Done():
			return nil, fmt.Errorf("Context cancelled")

		default:
			time.Sleep(5 * time.Second)

			out, err = svc.GetDocumentTextDetection(ctx, &textract.GetDocumentTextDetectionInput{JobId: resp.JobId})
			if err != nil {
				return nil, fmt.Errorf("getting textract job status: %w", err)
			}

			if out.JobStatus == types.JobStatusSucceeded || out.JobStatus == types.JobStatusFailed {
				break outer
			}
		}
	}

	var extracted string
	for _, block := range out.Blocks {
		if block.BlockType == types.BlockTypeLine && block.Text != nil {
			extracted += *block.Text
		}
	}

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
						Type: "text",
						Text: extracted,
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

	client := http.DefaultClient

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error making request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
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

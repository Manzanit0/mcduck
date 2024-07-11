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
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
	"github.com/martoche/pdf"
	"github.com/segmentio/ksuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
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

func trimMarkdownWrapper(s string) string {
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimSuffix(s, "```")
	return s
}

func doOpenAIRequest(ctx context.Context, request Request, openaiToken string) (*Receipt, error) {
	tp := otel.GetTracerProvider().Tracer("parser")
	ctx, span := tp.Start(ctx, "OpenAI: Prompt Chat Completion")
	defer span.End()

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %w", err)
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

type ReceiptParser interface {
	// PDF text extractor: passes the raw text
	// AWS Textract: passes the bytes of the PDF
	// OpenAI Vision: passes the bytes of image
	ExtractReceipt(context.Context, []byte) (*Receipt, error)
}

// TextractParser is a general-purpouse receipt parser that can process any
// kind of document by relying on AWS Textract. It'll then feed Textract's
// output to ChatGPT.
type TextractParser struct {
	openaiToken string
	tx          *textract.Client
	sthree      *s3.Client
	tp          trace.Tracer
}

var _ ReceiptParser = (*TextractParser)(nil)

func NewTextractParser(config aws.Config, openaiToken string) *TextractParser {
	tp := otel.GetTracerProvider().Tracer("parser")
	return &TextractParser{
		openaiToken: openaiToken,
		sthree:      s3.NewFromConfig(config),
		tx:          textract.NewFromConfig(config),
		tp:          tp,
	}
}

func (p TextractParser) ExtractReceipt(ctx context.Context, data []byte) (*Receipt, error) {
	ctx, span := p.tp.Start(ctx, "Extract Receipt: Textract")
	defer span.End()

	jobID, err := p.StartDocumentTextDetection(ctx, data)
	if err != nil {
		return nil, err
	}

	receiptText, err := p.GetDocumentText(ctx, jobID)
	if err != nil {
		return nil, err
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
						Text: receiptText,
					},
				},
			},
		},
	}

	return doOpenAIRequest(ctx, payload, p.openaiToken)
}

func (p TextractParser) StartDocumentTextDetection(ctx context.Context, data []byte) (string, error) {
	filename := fmt.Sprintf("%s.pdf", ksuid.New().String())

	_, span := p.tp.Start(ctx, "AWS S3 PUT")
	_, err := p.sthree.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("scratch-go"),
		Key:    aws.String(filename),
		Body:   bytes.NewBuffer(data),
	})
	if err != nil {
		return "", fmt.Errorf("AWS S3 PUT: %w", err)
	}
	span.End()

	_, span = p.tp.Start(ctx, "AWS Textract StartDocumentTextDetection")
	defer span.End()

	resp, err := p.tx.StartDocumentTextDetection(ctx, &textract.StartDocumentTextDetectionInput{
		DocumentLocation: &types.DocumentLocation{
			S3Object: &types.S3Object{
				Bucket: aws.String("scratch-go"),
				Name:   aws.String(filename),
			},
		},
		JobTag: aws.String("scratch-go"),
	})
	if err != nil {
		return "", fmt.Errorf("AWS Textract StartDocumentTextDetection: %w", err)
	}

	return *resp.JobId, nil
}

func (p TextractParser) GetDocumentText(ctx context.Context, jobID string) (string, error) {
	ctx, span := p.tp.Start(ctx, "Poll AWS Textract Results")
	defer span.End()

	var out *textract.GetDocumentTextDetectionOutput
	var err error
outer:
	for {
		select {

		case <-ctx.Done():
			return "", ctx.Err()

		default:
			_, span := p.tp.Start(ctx, "AWS Textract GetDocumentTextDetection")
			out, err = p.tx.GetDocumentTextDetection(ctx, &textract.GetDocumentTextDetectionInput{JobId: &jobID})
			if err != nil {
				return "", fmt.Errorf("getting textract job status: %w", err)
			}

			span.End()

			if out.JobStatus == types.JobStatusSucceeded || out.JobStatus == types.JobStatusFailed {
				break outer
			}

			_, span = p.tp.Start(ctx, "time.Sleep")
			time.Sleep(1 * time.Second)
			span.End()
		}
	}

	var extracted string
	for _, block := range out.Blocks {
		if block.BlockType == types.BlockTypeLine && block.Text != nil {
			extracted += *block.Text
		}
	}

	return extracted, nil
}

// AIVisionParser relies on OpenAI Vision to OCR receipts in image formats that
// are then fed into chatGPT.
type AIVisionParser struct {
	openaiToken string
}

func NewAIVisionParser(openaiToken string) *AIVisionParser {
	return &AIVisionParser{openaiToken: openaiToken}
}

var _ ReceiptParser = (*AIVisionParser)(nil)

func (p AIVisionParser) ExtractReceipt(ctx context.Context, data []byte) (*Receipt, error) {
	base64Image := base64.StdEncoding.EncodeToString(data)

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

	return doOpenAIRequest(ctx, payload, p.openaiToken)
}

// NaivePDFParser simply attempts to read the text from the PDF and pass it to
// openAI.
type NaivePDFParser struct {
	openaiToken string
	tp          trace.Tracer
}

func NewNaivePDFParser(openaiToken string) *NaivePDFParser {
	tp := otel.GetTracerProvider().Tracer("parser")
	return &NaivePDFParser{openaiToken: openaiToken, tp: tp}
}

var _ ReceiptParser = (*NaivePDFParser)(nil)

func (p NaivePDFParser) ExtractReceipt(ctx context.Context, data []byte) (*Receipt, error) {
	ctx, span := p.tp.Start(ctx, "Extract Receipt: Naive PDF read")
	defer span.End()

	_, span = p.tp.Start(ctx, "Extract Text from PDF")
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("new pdf reader: %w", err)
	}

	reader, err := r.GetPlainText()
	if err != nil {
		return nil, fmt.Errorf("get pdf text: %w", err)
	}

	buf, ok := reader.(*bytes.Buffer)
	if !ok {
		return nil, fmt.Errorf("github.com/martoche/pdf no longer uses bytes.Buffer to implement io.Reader")
	}

	extractedText := buf.String()
	span.End()

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
						Text: extractedText,
					},
				},
			},
		},
	}

	return doOpenAIRequest(ctx, payload, p.openaiToken)
}

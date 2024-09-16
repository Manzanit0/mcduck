package parser

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/manzanit0/mcduck/pkg/openai"
	"github.com/manzanit0/mcduck/pkg/xtrace"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
	"github.com/martoche/pdf"
	"github.com/segmentio/ksuid"
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

type Receipt struct {
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Description  string  `json:"description"`
	Vendor       string  `json:"vendor"`
	PurchaseDate string  `json:"purchase_date"`
}

type ReceiptParser interface {
	// PDF text extractor: passes the raw text
	// AWS Textract: passes the bytes of the PDF
	// OpenAI Vision: passes the bytes of image
	ExtractReceipt(context.Context, []byte) (*Receipt, *openai.Response, error)
}

// TextractParser is a general-purpouse receipt parser that can process any
// kind of document by relying on AWS Textract. It'll then feed Textract's
// output to ChatGPT.
type TextractParser struct {
	openaiToken string
	tx          *textract.Client
	sthree      *s3.Client
}

var _ ReceiptParser = (*TextractParser)(nil)

func NewTextractParser(config aws.Config, openaiToken string) *TextractParser {
	return &TextractParser{
		openaiToken: openaiToken,
		sthree:      s3.NewFromConfig(config),
		tx:          textract.NewFromConfig(config),
	}
}

func (p TextractParser) ExtractReceipt(ctx context.Context, data []byte) (*Receipt, *openai.Response, error) {
	ctx, span := xtrace.StartSpan(ctx, "Extract Receipt: Textract")
	defer span.End()

	jobID, err := p.StartDocumentTextDetection(ctx, data)
	if err != nil {
		span.RecordError(err)
		return nil, nil, err
	}

	receiptText, err := p.GetDocumentText(ctx, jobID)
	if err != nil {
		span.RecordError(err)
		return nil, nil, err
	}

	payload := openai.Request{
		Model:     "gpt-4o",
		MaxTokens: 300,
		Messages: []openai.Messages{
			{
				Role: "user",
				Content: []openai.Content{
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

	response, err := openai.Completions(ctx, p.openaiToken, payload)
	if err != nil {
		span.RecordError(err)
		return nil, response, fmt.Errorf("get openai completions: %w", err)
	}

	return receiptFromResponse(response)
}

func (p TextractParser) StartDocumentTextDetection(ctx context.Context, data []byte) (string, error) {
	filename := fmt.Sprintf("%s.pdf", ksuid.New().String())

	_, span := xtrace.StartSpan(ctx, "AWS S3 PUT")
	_, err := p.sthree.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("scratch-go"),
		Key:    aws.String(filename),
		Body:   bytes.NewBuffer(data),
	})
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("AWS S3 PUT: %w", err)
	}
	span.End()

	_, span = xtrace.StartSpan(ctx, "AWS Textract StartDocumentTextDetection")
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
		span.RecordError(err)
		return "", fmt.Errorf("AWS Textract StartDocumentTextDetection: %w", err)
	}

	return *resp.JobId, nil
}

func (p TextractParser) GetDocumentText(ctx context.Context, jobID string) (string, error) {
	ctx, span := xtrace.StartSpan(ctx, "Poll AWS Textract Results")
	defer span.End()

	var out *textract.GetDocumentTextDetectionOutput
	var err error
outer:
	for {
		select {

		case <-ctx.Done():
			return "", ctx.Err()

		default:
			_, span := xtrace.StartSpan(ctx, "AWS Textract GetDocumentTextDetection")
			out, err = p.tx.GetDocumentTextDetection(ctx, &textract.GetDocumentTextDetectionInput{JobId: &jobID})
			if err != nil {
				span.RecordError(err)
				return "", fmt.Errorf("getting textract job status: %w", err)
			}

			span.End()

			if out.JobStatus == types.JobStatusSucceeded || out.JobStatus == types.JobStatusFailed {
				break outer
			}

			_, span = xtrace.StartSpan(ctx, "time.Sleep")
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

func (p AIVisionParser) ExtractReceipt(ctx context.Context, data []byte) (*Receipt, *openai.Response, error) {
	base64Image := base64.StdEncoding.EncodeToString(data)

	payload := openai.Request{
		Model:     "gpt-4o",
		MaxTokens: 300,
		Messages: []openai.Messages{
			{
				Role: "user",
				Content: []openai.Content{
					{
						Type: "text",
						Text: initialPrompt,
					},
					{
						Type: "image_url",
						ImageURL: openai.ImageURL{
							URL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
						},
					},
				},
			},
		},
	}

	response, err := openai.Completions(ctx, p.openaiToken, payload)
	if err != nil {
		return nil, response, fmt.Errorf("get openai completions: %w", err)
	}

	return receiptFromResponse(response)
}

// NaivePDFParser simply attempts to read the text from the PDF and pass it to
// openAI.
type NaivePDFParser struct {
	openaiToken string
}

func NewNaivePDFParser(openaiToken string) *NaivePDFParser {
	return &NaivePDFParser{openaiToken: openaiToken}
}

var _ ReceiptParser = (*NaivePDFParser)(nil)

func (p NaivePDFParser) ExtractReceipt(ctx context.Context, data []byte) (*Receipt, *openai.Response, error) {
	ctx, span := xtrace.StartSpan(ctx, "Extract Receipt: Naive PDF read")
	defer span.End()

	_, span = xtrace.StartSpan(ctx, "Extract Text from PDF")
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		span.RecordError(err)
		return nil, nil, fmt.Errorf("new pdf reader: %w", err)
	}

	reader, err := r.GetPlainText()
	if err != nil {
		span.RecordError(err)
		return nil, nil, fmt.Errorf("get pdf text: %w", err)
	}

	buf, ok := reader.(*bytes.Buffer)
	if !ok {
		span.RecordError(err)
		return nil, nil, fmt.Errorf("github.com/martoche/pdf no longer uses bytes.Buffer to implement io.Reader")
	}

	extractedText := buf.String()
	span.End()

	payload := openai.Request{
		Model:     "gpt-4o",
		MaxTokens: 300,
		Messages: []openai.Messages{
			{
				Role: "user",
				Content: []openai.Content{
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

	response, err := openai.Completions(ctx, p.openaiToken, payload)
	if err != nil {
		return nil, response, fmt.Errorf("get openai completions: %w", err)
	}

	return receiptFromResponse(response)
}

func receiptFromResponse(response *openai.Response) (*Receipt, *openai.Response, error) {
	if len(response.Choices) == 0 {
		return nil, response, fmt.Errorf("openAI response has no choices")
	}

	j := trimMarkdownWrapper(response.Choices[0].Message.Content)

	var receipt Receipt
	err := json.Unmarshal([]byte(j), &receipt)
	if err != nil {
		return nil, response, fmt.Errorf("unmarshal receipt: %w", err)
	}

	return &receipt, response, nil
}

func trimMarkdownWrapper(s string) string {
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimSuffix(s, "```")
	return s
}

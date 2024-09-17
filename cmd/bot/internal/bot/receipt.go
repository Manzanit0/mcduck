package bot

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	receiptsv1 "github.com/manzanit0/mcduck/api/receipts.v1"
	"github.com/manzanit0/mcduck/api/receipts.v1/receiptsv1connect"
	usersv1 "github.com/manzanit0/mcduck/api/users.v1"
	"github.com/manzanit0/mcduck/api/users.v1/usersv1connect"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/manzanit0/mcduck/pkg/xtrace"
	"github.com/olekukonko/tablewriter"
	"go.opentelemetry.io/otel/codes"
)

const (
	defaultCurrency = "€"
)

func GetDocument(ctx context.Context, tgramClient tgram.Client, fileID string) ([]byte, error) {
	_, span := xtrace.StartSpan(ctx, "telegram.GetFile")
	file, err := tgramClient.GetFile(tgram.GetFileRequest{FileID: fileID})
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("get file: %w", err)
	}
	span.End()

	_, span = xtrace.StartSpan(ctx, "telegram.DownloadFile")
	fileData, err := tgramClient.DownloadFile(file)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("download file: %w", err)
	}
	span.End()

	return fileData, nil
}

func ParseReceipt(ctx context.Context, tgramClient tgram.Client, usersClient usersv1connect.UsersServiceClient, receiptsClient receiptsv1connect.ReceiptsServiceClient, r *tgram.WebhookRequest) *tgram.WebhookResponse {
	ctx, span := xtrace.StartSpan(ctx, "Parse Receipt")
	defer span.End()

	var fileID string
	var fileSize int64

	if r.Message.Document != nil {
		fileID = r.Message.Document.FileID
	} else if len(r.Message.Photos) > 0 {
		// Get the biggest photo: this will ensure better parsing by parser service.
		for _, p := range r.Message.Photos {
			if p.FileSize != nil && *p.FileSize > fileSize {
				fileID = p.FileID
				fileSize = *p.FileSize
			}
		}
	}

	fileData, err := GetDocument(ctx, tgramClient, fileID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "tgram.DownloadFile:", "error", err.Error())
		return tgram.NewHTMLResponse(fmt.Sprintf("unable to download file from Telegram servers: %s", err.Error()), r.GetFromID())
	}

	if len(fileData) == 0 {
		return tgram.NewHTMLResponse("empty file", r.GetFromID())
	}

	getUserReq := connect.Request[usersv1.GetUserRequest]{
		Msg: &usersv1.GetUserRequest{
			TelegramChatId: int64(r.GetFromID()),
		},
	}

	// FIXME: this is a bit of a hack - since user validity isn't validated
	// against the database, this API call should work. This hack, however, won't
	// work for other requests that use the user to fetch user-bound data.
	onBehalfOf := "bot@mcduck.com"
	token, err := auth.GenerateJWT(onBehalfOf)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Generate JWT:", "error", err.Error())
		return tgram.NewHTMLResponse(fmt.Sprintf("generate JWT: %s", err.Error()), r.GetFromID())
	}

	getUserReq.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := usersClient.GetUser(ctx, &getUserReq)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "mcduck.SearchUserByChatID:", "error", err.Error())
		return tgram.NewHTMLResponse(fmt.Sprintf("unable to find user: %s", err.Error()), r.GetFromID())
	}

	createReceiptReq := connect.Request[receiptsv1.CreateReceiptsRequest]{
		Msg: &receiptsv1.CreateReceiptsRequest{
			ReceiptFiles: [][]byte{fileData},
		},
	}

	onBehalfOf = resp.Msg.User.Email
	token, err = auth.GenerateJWT(onBehalfOf)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Generate JWT:", "error", err.Error())
		return tgram.NewHTMLResponse(fmt.Sprintf("generate JWT: %s", err.Error()), r.GetFromID())
	}

	createReceiptReq.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := receiptsClient.CreateReceipts(ctx, &createReceiptReq)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "CreateReceipt", "error", err.Error())
		return tgram.NewHTMLResponse(fmt.Sprintf("unable to parser receipt: %s", err.Error()), r.GetFromID())
	}

	return tgram.NewMarkdownResponse(newBreakdownTgramMessage(map[string]float64{
		res.Msg.Receipts[0].Expenses[0].Description: float64(expense.ConvertToDollar(int32(res.Msg.Receipts[0].Expenses[0].Amount))),
	}), r.GetFromID())
}

func newBreakdownTgramMessage(amounts map[string]float64) string {
	b := bytes.NewBuffer([]byte{})
	table := tablewriter.NewWriter(b)

	table.SetHeader([]string{"Item", "Amount"})

	var total float64
	for k, v := range amounts {
		// We're trimming the item name because we want the table to render
		// properly on small phones. The reference is an iPhone SE.
		item := strings.TrimSpace(strings.Title(strings.ToLower(fmt.Sprintf("%.14s", k))))
		table.Append([]string{item, fmt.Sprintf("%.2f%s", v, defaultCurrency)})
		total += v
	}

	// table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowSeparator("-")
	table.SetAutoFormatHeaders(false)
	table.SetBorder(false)
	table.SetFooter([]string{"TOTAL", fmt.Sprintf("%.2f%s", total, defaultCurrency)})

	table.Render()

	return fmt.Sprintf("```%s```", b.String())
}

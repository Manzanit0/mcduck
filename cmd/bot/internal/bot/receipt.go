package bot

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/olekukonko/tablewriter"
)

const (
	defaultCurrency = "€"
)

func ParseReceipt(ctx context.Context, tgramClient tgram.Client, mcduckClient client.McDuckClient, r *tgram.WebhookRequest) *tgram.WebhookResponse {
	// Get the biggest photo: this will ensure better parsing by parser service.
	var fileID string
	var fileSize int64
	for _, p := range r.Message.Photos {
		if p.FileSize != nil && *p.FileSize > fileSize {
			fileID = p.FileID
			fileSize = *p.FileSize
		}
	}

	file, err := tgramClient.GetFile(tgram.GetFileRequest{FileID: fileID})
	if err != nil {
		slog.Error("tgram.GetFile", "error", err.Error())
		return tgram.NewHTMLResponse(fmt.Sprintf("unable to get file from Telegram servers: %s", err.Error()), r.GetFromID())
	}

	fileData, err := tgramClient.DownloadFile(file)
	if err != nil {
		slog.Error("tgram.DownloadFile:", "error", err.Error())
		return tgram.NewHTMLResponse(fmt.Sprintf("unable to download file from Telegram servers: %s", err.Error()), r.GetFromID())
	}

	if len(fileData) == 0 {
		return tgram.NewHTMLResponse("empty file", r.GetFromID())
	}

	// FIXME: this is a bit of a hack - since user validity isn't validated
	// against the database, this API call should work. This hack, however, won't
	// work for other requests that use the user to fetch user-bound data.
	onBehalfOf := "bot@mcduck.com"
	resp, err := mcduckClient.SearchUserByChatID(ctx, onBehalfOf, r.GetFromID())
	if err != nil {
		slog.Error("mcduck.SearchUserByChatID:", "error", err.Error())
		return tgram.NewHTMLResponse(fmt.Sprintf("unable to find user: %s", err.Error()), r.GetFromID())
	}

	onBehalfOf = resp.User.Email
	res, err := mcduckClient.CreateReceipt(ctx, onBehalfOf, fileData)
	if err != nil {
		slog.Error("mcduck.CreateReceipt", "error", err.Error())
		return tgram.NewHTMLResponse(fmt.Sprintf("unable to parser receipt: %s", err.Error()), r.GetFromID())
	}

	return tgram.NewMarkdownResponse(newBreakdownTgramMessage(res.Amounts), r.GetFromID())
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

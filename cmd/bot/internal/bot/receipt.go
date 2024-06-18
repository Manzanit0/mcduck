package bot

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/pkg/invx"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/olekukonko/tablewriter"
)

const (
	defaultCurrency = "€"
)

func ParseReceipt(ctx context.Context, tgramClient tgram.Client, invxClient invx.Client, mcduckClient client.McDuckClient, r *tgram.WebhookRequest) *tgram.WebhookResponse {
	// Get the biggest photo: this will ensure better parsing by invx service.
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
		log.Println("[ERROR] tgram.GetFile:", err.Error())
		return tgram.NewMarkdownResponse(fmt.Sprintf("unable to get file from Telegram servers: %s", err.Error()), r.GetFromID())
	}

	fileData, err := tgramClient.DownloadFile(file)
	if err != nil {
		log.Println("[ERROR] tgram.DownloadFile:", err.Error())
		return tgram.NewMarkdownResponse(fmt.Sprintf("unable to download file from Telegram servers: %s", err.Error()), r.GetFromID())
	}

	if len(fileData) == 0 {
		return tgram.NewMarkdownResponse("empty file", r.GetFromID())
	}

	amounts, err := invxClient.ParseReceipt(ctx, fileData)
	if err != nil {
		log.Println("[ERROR] invx.ParseReceipt", err.Error())
		return tgram.NewMarkdownResponse(fmt.Sprintf("unable to parser receipt: %s", err.Error()), r.GetFromID())
	}

	_, err = mcduckClient.CreateReceipt(ctx, fileData)
	if err != nil {
		log.Println("[ERROR] mcduck.CreateReceipt", err.Error())
		return tgram.NewMarkdownResponse(fmt.Sprintf("unable to parser receipt: %s", err.Error()), r.GetFromID())
	}

	return tgram.NewMarkdownResponse(newBreakdownTgramMessage(amounts), r.GetFromID())
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

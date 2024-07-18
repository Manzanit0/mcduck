package bot

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/manzanit0/mcduck/pkg/xtrace"
)

func LoginLink(ctx context.Context, r *tgram.WebhookRequest) *tgram.WebhookResponse {
	_, span := xtrace.StartSpan(ctx, "Build Login Link")
	defer span.End()

	host := os.Getenv("MCDUCK_HOST")
	id := url.QueryEscape(fmt.Sprint(r.GetFromID()))
	link := fmt.Sprintf("https://%s/connect?tgram=%s", host, id)

	res := tgram.NewMarkdownResponse(fmt.Sprintf("To allow me access to your data, log in the portal: %s", link), r.GetFromID())
	res.ParseMode = tgram.ParseModeMarkdownV1
	return res
}

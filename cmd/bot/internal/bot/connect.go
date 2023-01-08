package bot

import (
	"fmt"
	"net/url"
	"os"

	"github.com/manzanit0/mcduck/pkg/tgram"
)

func LoginLink(r *tgram.WebhookRequest) *tgram.WebhookResponse {
	host := os.Getenv("MCDUCK_HOST")
	id := url.QueryEscape(fmt.Sprint(r.GetFromID()))
	link := fmt.Sprintf("https://%s/connect?tgram=%s", host, id)

	res := tgram.NewMarkdownResponse(fmt.Sprintf("To allow me access to your data, log in the portal: %s", link), r.GetFromID())
	res.ParseMode = tgram.ParseModeMarkdownV1
	return res
}

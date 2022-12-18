package tgram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type WebhookRequest struct {
	UpdateID      int      `json:"update_id"`
	Message       *Message `json:"message"`
	EditedMessage *Message `json:"edited_message"`
}

func (w WebhookRequest) GetFromUsername() string {
	if w.Message != nil {
		return w.Message.From.Username
	}

	if w.EditedMessage != nil {
		return w.EditedMessage.From.Username
	}

	return ""
}

func (w WebhookRequest) GetFromID() int {
	if w.Message != nil {
		return w.Message.From.ID
	}

	if w.EditedMessage != nil {
		return w.EditedMessage.From.ID
	}

	return 0
}

func (w WebhookRequest) GetFromFirstName() string {
	if w.Message != nil {
		return w.Message.From.FirstName
	}

	if w.EditedMessage != nil {
		return w.EditedMessage.From.FirstName
	}

	return ""
}

func (w WebhookRequest) GetFromLastName() string {
	if w.Message != nil {
		return w.Message.From.LastName
	}

	if w.EditedMessage != nil {
		return w.EditedMessage.From.LastName
	}

	return ""
}

func (w WebhookRequest) GetFromLanguageCode() string {
	if w.Message != nil {
		return w.Message.From.LanguageCode
	}

	if w.EditedMessage != nil {
		return w.EditedMessage.From.LanguageCode
	}

	return ""
}

type From struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type Chat struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	From      From   `json:"from"`
	Chat      Chat   `json:"chat"`
	Date      int    `json:"date"`
	Text      string `json:"text"`
}

func NewMessage(chatID, text string) Message {
	return Message{}
}

type Client interface {
	SendMessage(SendMessageRequest) error
}

type client struct {
	h        *http.Client
	botToken string
}

func NewClient(h *http.Client, token string) *client {
	return &client{h: h, botToken: token}
}

type SendMessageRequest struct {
	ChatID    int64     `json:"chat_id,omitempty"`
	Text      string    `json:"text,omitempty"`
	ParseMode ParseMode `json:"parse_mode,omitempty"`
}

type ParseMode string

const (
	ParseModeMarkdownV2 ParseMode = "MarkdownV2"
	ParseModeMarkdownV1 ParseMode = "Markdown"
	ParseModeHTML       ParseMode = "HTML"
)

func (c *client) SendMessage(m SendMessageRequest) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.botToken)

	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("unable to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("unable to create http req: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := c.h.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request: %w", err)
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("request failed with status %d, but unable to read body: %w", res.StatusCode, err)
	}

	return fmt.Errorf("request failed with status %d and body %s", res.StatusCode, string(data))
}

package tgram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type WebhookRequest struct {
	UpdateID      int      `json:"update_id"`
	Message       *Message `json:"message"`
	EditedMessage *Message `json:"edited_message"`
}

type WebhookResponse struct {
	ChatID    int       `json:"chat_id,omitempty"`
	Text      string    `json:"text,omitempty"`
	ParseMode ParseMode `json:"parse_mode,omitempty"`
	Method    string    `json:"method,omitempty"`
}

func NewMarkdownResponse(text string, chatID int) *WebhookResponse {
	return &WebhookResponse{
		ChatID:    chatID,
		Text:      text,
		Method:    "sendMessage",
		ParseMode: ParseModeMarkdownV2,
	}
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

type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

// Chat represents a chat.
type Chat struct {
	ID        int     `json:"id"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Username  *string `json:"username"`
	Type      string  `json:"type"` // can be of type "private", "group", "supergroup" or "channel"
}

// Message represents a message.
type Message struct {
	MessageID int         `json:"message_id"`
	From      *User       `json:"from"`
	Chat      Chat        `json:"chat"`
	Date      int         `json:"date"`
	Text      *string     `json:"text"`
	Document  *Document   `json:"document"`
	Photos    []PhotoSize `json:"photo"`
}

// Document represents a general file (as opposed to photos, voice messages and audio files).
type Document struct {
	FileID       string  `json:"file_id"`
	FileUniqueID string  `json:"file_unique_id"`
	FileName     *string `json:"file_name"`
	MimeType     *string `json:"mime_type"`
	FileSize     *int64  `json:"file_size"`
}

// PhotoSize represents one size of a photo or a file / sticker thumbnail.
type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     *int64 `json:"file_size"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

func NewMessage(chatID, text string) Message {
	return Message{}
}

type Client interface {
	SendMessage(SendMessageRequest) error
	GetFile(GetFileRequest) (*File, error)
	DownloadFile(*File) ([]byte, error)
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

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("request failed with status %d, but unable to read body: %w", res.StatusCode, err)
	}

	return fmt.Errorf("request failed with status %d and body %s", res.StatusCode, string(data))
}

type GetFileRequest struct {
	FileID string `json:"file_id"`
}

type GetFileResponse struct {
	Ok     bool `json:"ok"`
	Result File `json:"result"`
}

type File struct {
	ID       string `json:"file_id"`
	UniqueID string `json:"file_unique_id"`
	Size     int    `json:"file_size"`
	Path     string `json:"file_path"`
}

type ErrorResponse struct {
	OK          bool   `json:"ok"`
	Code        int    `json:"error_code"`
	Description string `json:"description"`
}

func (c *client) GetFile(m GetFileRequest) (*File, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getFile", c.botToken)

	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("create http req: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if res.StatusCode != 200 {
		var errResp ErrorResponse
		_ = json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("request failed: %s", errResp.Description)
	}

	var fileResponse GetFileResponse
	err = json.Unmarshal(data, &fileResponse)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response body: %w", err)
	}

	return &fileResponse.Result, nil
}

func (c *client) DownloadFile(file *File) ([]byte, error) {
	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.botToken, file.Path)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create http req: %w", err)
	}

	res, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if res.StatusCode != 200 {
		var errResp ErrorResponse
		_ = json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("request failed: %s", errResp.Description)
	}

	return data, nil
}

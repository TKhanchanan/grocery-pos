package line

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultPushURL = "https://api.line.me/v2/bot/message/push"

type Message interface {
	lineMessage()
}

type TextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (TextMessage) lineMessage() {}

type FlexMessage struct {
	Type     string `json:"type"`
	AltText  string `json:"altText"`
	Contents Bubble `json:"contents"`
}

func (FlexMessage) lineMessage() {}

type Client struct {
	httpClient *http.Client
	pushURL    string
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 6 * time.Second}
	}
	return &Client{
		httpClient: httpClient,
		pushURL:    defaultPushURL,
	}
}

func NewTextMessage(text string) TextMessage {
	return TextMessage{Type: "text", Text: text}
}

func (c *Client) Push(ctx context.Context, token, targetID string, messages ...Message) error {
	body, err := json.Marshal(struct {
		To       string    `json:"to"`
		Messages []Message `json:"messages"`
	}{
		To:       targetID,
		Messages: messages,
	})
	if err != nil {
		return fmt.Errorf("encode LINE push payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.pushURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create LINE push request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send LINE push request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	bodyBytes, _ := io.ReadAll(io.LimitReader(res.Body, 512))
	detail := strings.TrimSpace(string(bodyBytes))
	if detail == "" {
		detail = http.StatusText(res.StatusCode)
	}
	return fmt.Errorf("LINE API %d: %s", res.StatusCode, detail)
}

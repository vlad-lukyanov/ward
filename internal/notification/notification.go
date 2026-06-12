package notification

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

type Config struct {
	Enabled  bool           `yaml:"enabled"`
	Telegram TelegramConfig `yaml:"telegram"`
}

type Notifier struct {
	enabled  bool
	botToken string
	chatID   string
	logger   *log.Logger
	client   *http.Client
}

func New(enabled bool, botToken, chatID string, logger *log.Logger) *Notifier {
	return &Notifier{
		enabled:  enabled,
		botToken: botToken,
		chatID:   chatID,
		logger:   logger,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *Notifier) Send(service, message string) error {
	if !n.enabled || n.botToken == "" || n.chatID == "" {
		return nil
	}

	text := fmt.Sprintf("🛡 *WARD*\n\n*%s*\n%s", service, message)

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)

	data := url.Values{}
	data.Set("chat_id", n.chatID)
	data.Set("text", text)
	data.Set("parse_mode", "Markdown")

	resp, err := n.client.PostForm(apiURL, data)
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram api error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (n *Notifier) NotifyRestart(service string, attempt, max int) {
	msg := fmt.Sprintf("Restarted (attempt %d/%d)", attempt, max)
	if err := n.Send(service, msg); err != nil {
		n.logger.Printf("failed to send notification: %v", err)
	}
}

func (n *Notifier) NotifyMaxRestarts(service string, max int) {
	msg := fmt.Sprintf("Exceeded max restarts (%d), no longer restarting", max)
	if err := n.Send(service, msg); err != nil {
		n.logger.Printf("failed to send notification: %v", err)
	}
}

func (n *Notifier) NotifyFailedRestart(service string, err error) {
	msg := fmt.Sprintf("Restart failed: %v", err)
	if err := n.Send(service, msg); err != nil {
		n.logger.Printf("failed to send notification: %v", err)
	}
}

func (n *Notifier) Reload(enabled bool, botToken, chatID string) {
	n.enabled = enabled
	n.botToken = botToken
	n.chatID = chatID
}

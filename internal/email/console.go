package email

import (
	"context"
	"log/slog"
)

type ConsoleSender struct{}

func NewConsoleSender() *ConsoleSender { return &ConsoleSender{} }

func (s *ConsoleSender) Send(_ context.Context, msg Message) error {
	slog.Info("EMAIL",
		"to", msg.To,
		"subject", msg.Subject,
		"body", msg.Body,
	)
	return nil
}

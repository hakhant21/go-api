package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/hakhant21/go-starter/internal/email"
	"github.com/hakhant21/go-starter/internal/model"
	tokenrepo "github.com/hakhant21/go-starter/internal/repository/token"
	"github.com/hakhant21/go-starter/pkg/jwt"
)

const (
	verifyTTL = 24 * time.Hour
	resetTTL  = 1 * time.Hour
)

type Notifier struct {
	tokenRepo tokenrepo.Repository
	mailer    email.Sender
	appURL    string
}

func NewNotifier(tr tokenrepo.Repository, m email.Sender, appURL string) *Notifier {
	return &Notifier{tokenRepo: tr, mailer: m, appURL: appURL}
}

func (n *Notifier) SendVerification(ctx context.Context, u *model.User) error {
	raw, hash, err := jwt.GenerateRefreshToken()
	if err != nil {
		return err
	}
	_ = n.tokenRepo.InvalidateAllForUser(ctx, u.ID, model.TokenTypeEmailVerify)
	tok := &model.OneTimeToken{
		UserID:    u.ID,
		TokenHash: hash,
		Type:      model.TokenTypeEmailVerify,
		ExpiresAt: time.Now().Add(verifyTTL),
	}
	if err := n.tokenRepo.Create(ctx, tok); err != nil {
		return err
	}
	link := fmt.Sprintf("%s/verify-email?token=%s", n.appURL, raw)
	return n.mailer.Send(ctx, email.Message{
		To:      u.Email,
		Subject: "Verify your email",
		Body:    fmt.Sprintf("Hi %s,\n\nVerify your email by visiting:\n%s\n\nLink expires in 24h.", u.Name, link),
	})
}

func (n *Notifier) SendPasswordReset(ctx context.Context, u *model.User) error {
	raw, hash, err := jwt.GenerateRefreshToken()
	if err != nil {
		return err
	}
	_ = n.tokenRepo.InvalidateAllForUser(ctx, u.ID, model.TokenTypePasswordReset)
	tok := &model.OneTimeToken{
		UserID:    u.ID,
		TokenHash: hash,
		Type:      model.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(resetTTL),
	}
	if err := n.tokenRepo.Create(ctx, tok); err != nil {
		return err
	}
	link := fmt.Sprintf("%s/reset-password?token=%s", n.appURL, raw)
	return n.mailer.Send(ctx, email.Message{
		To:      u.Email,
		Subject: "Reset your password",
		Body:    fmt.Sprintf("Hi %s,\n\nReset your password by visiting:\n%s\n\nLink expires in 1h.", u.Name, link),
	})
}

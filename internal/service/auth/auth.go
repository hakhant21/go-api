package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/hakhant21/go-starter/internal/dto"
	"github.com/hakhant21/go-starter/internal/email"
	"github.com/hakhant21/go-starter/internal/model"
	"github.com/hakhant21/go-starter/internal/repository"
	refreshtokenrepo "github.com/hakhant21/go-starter/internal/repository/refresh_token"
	tokenrepo "github.com/hakhant21/go-starter/internal/repository/token"
	userrepo "github.com/hakhant21/go-starter/internal/repository/user"
	"github.com/hakhant21/go-starter/internal/service"
	"github.com/hakhant21/go-starter/pkg/jwt"
)

type Service struct {
	userRepo     userrepo.Repository
	refreshRepo  refreshtokenrepo.Repository
	tokenRepo    tokenrepo.Repository
	notifier     *Notifier
	secret       string
	accessExpiry int64
	refreshTTL   time.Duration
}

func New(
	userRepo userrepo.Repository,
	refreshRepo refreshtokenrepo.Repository,
	tokenRepo tokenrepo.Repository,
	mailer email.Sender,
	secret string,
	accessExpirySeconds int64,
	refreshTTL time.Duration,
	appBaseURL string,
) *Service {
	return &Service{
		userRepo:     userRepo,
		refreshRepo:  refreshRepo,
		tokenRepo:    tokenRepo,
		notifier:     NewNotifier(tokenRepo, mailer, appBaseURL),
		secret:       secret,
		accessExpiry: accessExpirySeconds,
		refreshTTL:   refreshTTL,
	}
}

func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	u, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, service.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, service.ErrUnauthorized
	}
	if !u.Active {
		return nil, service.ErrUnauthorized
	}
	if !u.EmailVerified {
		return nil, service.ErrEmailNotVerified
	}
	return s.issueTokens(ctx, u)
}

func (s *Service) Refresh(ctx context.Context, raw string) (*dto.AuthResponse, error) {
	hash := jwt.HashToken(raw)
	rt, err := s.refreshRepo.GetByHash(ctx, hash)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, service.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	if rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return nil, service.ErrUnauthorized
	}
	if err := s.refreshRepo.Revoke(ctx, hash); err != nil {
		return nil, err
	}
	u, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}
	if !u.Active {
		return nil, service.ErrUnauthorized
	}
	return s.issueTokens(ctx, u)
}

func (s *Service) Logout(ctx context.Context, raw string) error {
	return s.refreshRepo.Revoke(ctx, jwt.HashToken(raw))
}

func (s *Service) VerifyEmail(ctx context.Context, raw string) error {
	tok, err := s.tokenRepo.GetValid(ctx, jwt.HashToken(raw), model.TokenTypeEmailVerify)
	if err != nil {
		return service.ErrTokenInvalid
	}
	u, err := s.userRepo.GetByID(ctx, tok.UserID)
	if err != nil {
		return err
	}
	if u.EmailVerified {
		return service.ErrEmailAlreadyVerified
	}
	if err := s.userRepo.SetEmailVerified(ctx, u.ID, time.Now()); err != nil {
		return err
	}
	return s.tokenRepo.MarkUsed(ctx, tok.ID)
}

func (s *Service) ResendVerification(ctx context.Context, emailAddr string) error {
	u, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(emailAddr)))
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if u.EmailVerified {
		return nil
	}
	return s.notifier.SendVerification(ctx, u)
}

func (s *Service) ForgotPassword(ctx context.Context, emailAddr string) error {
	u, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(emailAddr)))
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.notifier.SendPasswordReset(ctx, u)
}

func (s *Service) ResetPassword(ctx context.Context, raw, newPassword string) error {
	tok, err := s.tokenRepo.GetValid(ctx, jwt.HashToken(raw), model.TokenTypePasswordReset)
	if err != nil {
		return service.ErrTokenInvalid
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.userRepo.UpdatePassword(ctx, tok.UserID, string(hashed)); err != nil {
		return err
	}
	if err := s.tokenRepo.MarkUsed(ctx, tok.ID); err != nil {
		return err
	}
	return s.refreshRepo.RevokeAllForUser(ctx, tok.UserID)
}

func (s *Service) issueTokens(ctx context.Context, u *model.User) (*dto.AuthResponse, error) {
	access, err := jwt.Generate(u.ID, u.Email, s.secret, s.accessExpiry)
	if err != nil {
		return nil, err
	}
	raw, hash, err := jwt.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	rt := &model.RefreshToken{UserID: u.ID, TokenHash: hash, ExpiresAt: time.Now().Add(s.refreshTTL)}
	if err := s.refreshRepo.Create(ctx, rt); err != nil {
		return nil, err
	}
	return &dto.AuthResponse{
		Token:        access,
		RefreshToken: raw,
		User:         dto.UserToResponse(u),
	}, nil
}

func (s *Service) SendVerificationEmail(ctx context.Context, u *model.User) error {
	return s.notifier.SendVerification(ctx, u)
}

package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/hakhant21/go-starter/internal/model"
	"github.com/hakhant21/go-starter/internal/repository"
)

type Repository interface {
	Create(ctx context.Context, t *model.OneTimeToken) error
	GetValid(ctx context.Context, hash string, typ model.TokenType) (*model.OneTimeToken, error)
	MarkUsed(ctx context.Context, id uint) error
	InvalidateAllForUser(ctx context.Context, userID uint, typ model.TokenType) error
	DeleteExpired(ctx context.Context) error
}

type repo struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) Create(ctx context.Context, t *model.OneTimeToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *repo) GetValid(ctx context.Context, hash string, typ model.TokenType) (*model.OneTimeToken, error) {
	var t model.OneTimeToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND type = ? AND used_at IS NULL AND expires_at > ?", hash, typ, time.Now()).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	return &t, nil
}

func (r *repo) MarkUsed(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.OneTimeToken{}).Where("id = ?", id).
		Update("used_at", time.Now()).Error
}

func (r *repo) InvalidateAllForUser(ctx context.Context, userID uint, typ model.TokenType) error {
	return r.db.WithContext(ctx).Model(&model.OneTimeToken{}).
		Where("user_id = ? AND type = ? AND used_at IS NULL", userID, typ).
		Update("used_at", time.Now()).Error
}

func (r *repo) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&model.OneTimeToken{}).Error
}

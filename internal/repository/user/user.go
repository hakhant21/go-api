package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/hakhant21/go-starter/internal/model"
	"github.com/hakhant21/go-starter/internal/repository"
)

type Filter struct {
	Search   string
	RoleName string
	Active   *bool
	Limit    int
	Offset   int
}

type Repository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, limit, offset int) ([]model.User, int64, error)
	ListFiltered(ctx context.Context, f Filter) ([]model.User, int64, error)
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id uint) error
	SetActive(ctx context.Context, id uint, active bool) error
	SetEmailVerified(ctx context.Context, userID uint, verifiedAt time.Time) error
	UpdatePassword(ctx context.Context, userID uint, passwordHash string) error
}

type repo struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *repo) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}
	return &u, nil
}

func (r *repo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *repo) List(ctx context.Context, limit, offset int) ([]model.User, int64, error) {
	var (
		users []model.User
		total int64
	)
	q := r.db.WithContext(ctx).Model(&model.User{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *repo) ListFiltered(ctx context.Context, f Filter) ([]model.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.User{})
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("name ILIKE ? OR email ILIKE ?", like, like)
	}
	if f.Active != nil {
		q = q.Where("active = ?", *f.Active)
	}
	if f.RoleName != "" {
		q = q.Joins("JOIN user_roles ur ON ur.user_id = users.id").
			Joins("JOIN roles r ON r.id = ur.role_id").
			Where("r.name = ?", f.RoleName).
			Distinct("users.*")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	if err := q.Order("users.created_at DESC").Limit(f.Limit).Offset(f.Offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *repo) Update(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *repo) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.User{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *repo) SetActive(ctx context.Context, id uint, active bool) error {
	res := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("active", active)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *repo) SetEmailVerified(ctx context.Context, userID uint, verifiedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]any{"email_verified": true, "email_verified_at": verifiedAt}).Error
}

func (r *repo) UpdatePassword(ctx context.Context, userID uint, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update("password", passwordHash).Error
}

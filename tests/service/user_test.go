package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hakhant21/go-starter/internal/dto"
	"github.com/hakhant21/go-starter/internal/email"
	refreshtokenrepo "github.com/hakhant21/go-starter/internal/repository/refresh_token"
	rbacrepo "github.com/hakhant21/go-starter/internal/repository/rbac"
	tokenrepo "github.com/hakhant21/go-starter/internal/repository/token"
	userrepo "github.com/hakhant21/go-starter/internal/repository/user"
	"github.com/hakhant21/go-starter/internal/service"
	authsvc "github.com/hakhant21/go-starter/internal/service/auth"
	rbacsvc "github.com/hakhant21/go-starter/internal/service/rbac"
	usersvc "github.com/hakhant21/go-starter/internal/service/user"
	"github.com/hakhant21/go-starter/internal/testutil"
)

type noopMailer struct{}

func (n *noopMailer) Send(_ context.Context, _ email.Message) error { return nil }

func TestUserService_Create_DuplicateEmail(t *testing.T) {
	db := testutil.NewTestDB(t)
	userRepo := userrepo.New(db)
	refreshRepo := refreshtokenrepo.New(db)
	tokenRepo := tokenrepo.New(db)
	rbacRepo := rbacrepo.New(db)

	rbacSvc := rbacsvc.New(rbacRepo, nil, nil)
	authSvc := authsvc.New(userRepo, refreshRepo, tokenRepo, &noopMailer{},
		"secret", 3600, 24*3600, "http://x")
	userSvc := usersvc.New(userRepo, authSvc, rbacSvc, nil)

	ctx := context.Background()
	req := dto.CreateUserRequest{Name: "Alice", Email: "a@x.com", Password: "password123"}
	if _, err := userSvc.Create(ctx, req); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := userSvc.Create(ctx, req)
	if !errors.Is(err, service.ErrEmailTaken) {
		t.Fatalf("want ErrEmailTaken, got %v", err)
	}
}

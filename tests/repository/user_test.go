package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hakhant21/go-starter/internal/model"
	"github.com/hakhant21/go-starter/internal/repository"
	userrepo "github.com/hakhant21/go-starter/internal/repository/user"
	"github.com/hakhant21/go-starter/internal/testutil"
)

func TestUserRepository_CreateAndGet(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := userrepo.New(db)
	ctx := context.Background()

	u := &model.User{Name: "Alice", Email: "alice@x.com", Password: "hash", Active: true}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("expected ID")
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Email != "alice@x.com" {
		t.Errorf("want alice@x.com, got %s", got.Email)
	}
}

func TestUserRepository_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := userrepo.New(db)

	_, err := repo.GetByID(context.Background(), 9999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

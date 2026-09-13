package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hakhant21/go-starter/internal/model"
)

const (
	userBaseTTL = 5 * time.Minute
	userJitter  = 30 * time.Second
)

type UserCache struct{ rdb *redis.Client }

func NewUserCache(rdb *redis.Client) *UserCache { return &UserCache{rdb: rdb} }

func (c *UserCache) key(id uint) string { return fmt.Sprintf("user:%d", id) }

func (c *UserCache) Get(ctx context.Context, id uint) (*model.User, bool) {
	raw, err := c.rdb.Get(ctx, c.key(id)).Bytes()
	if err != nil {
		return nil, false
	}
	var u model.User
	if err := json.Unmarshal(raw, &u); err != nil {
		return nil, false
	}
	return &u, true
}

func (c *UserCache) Set(ctx context.Context, u *model.User) {
	raw, _ := json.Marshal(u)
	ttl := userBaseTTL + time.Duration(rand.Int64N(int64(userJitter)))
	_ = c.rdb.Set(ctx, c.key(u.ID), raw, ttl).Err()
}

func (c *UserCache) Invalidate(ctx context.Context, id uint) {
	_ = c.rdb.Del(ctx, c.key(id)).Err()
}

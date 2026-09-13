package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	rbacPermBaseTTL = 60 * time.Second
	rbacPermJitter  = 20 * time.Second
)

type RBACCache struct{ rdb *redis.Client }

func NewRBACCache(rdb *redis.Client) *RBACCache { return &RBACCache{rdb: rdb} }

func (c *RBACCache) key(userID uint) string { return fmt.Sprintf("rbac:perms:%d", userID) }

func (c *RBACCache) Get(ctx context.Context, userID uint) ([]string, bool) {
	raw, err := c.rdb.Get(ctx, c.key(userID)).Bytes()
	if err != nil {
		return nil, false
	}
	var perms []string
	if err := json.Unmarshal(raw, &perms); err != nil {
		return nil, false
	}
	return perms, true
}

func (c *RBACCache) Set(ctx context.Context, userID uint, perms []string) {
	raw, _ := json.Marshal(perms)
	ttl := rbacPermBaseTTL + time.Duration(rand.Int64N(int64(rbacPermJitter)))
	_ = c.rdb.Set(ctx, c.key(userID), raw, ttl).Err()
}

func (c *RBACCache) Invalidate(ctx context.Context, userID uint) {
	_ = c.rdb.Del(ctx, c.key(userID)).Err()
}

func (c *RBACCache) InvalidateUsers(ctx context.Context, userIDs []uint) {
	if len(userIDs) == 0 {
		return
	}
	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = c.key(id)
	}
	_ = c.rdb.Del(ctx, keys...).Err()
}

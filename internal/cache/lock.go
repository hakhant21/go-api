package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLockTimeout = errors.New("lock acquisition timeout")

type Lock struct{ rdb *redis.Client }

func NewLock(rdb *redis.Client) *Lock { return &Lock{rdb: rdb} }

func (l *Lock) Acquire(ctx context.Context, key string, ttl, wait time.Duration) (func(), error) {
	token := randomToken()
	deadline := time.Now().Add(wait)

	for {
		ok, err := l.rdb.SetNX(ctx, key, token, ttl).Result()
		if err != nil {
			return nil, err
		}
		if ok {
			return func() { l.release(context.Background(), key, token) }, nil
		}
		if time.Now().After(deadline) {
			return nil, ErrLockTimeout
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func (l *Lock) release(ctx context.Context, key, token string) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end`
	_ = l.rdb.Eval(ctx, script, []string{key}, token).Err()
}

func randomToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

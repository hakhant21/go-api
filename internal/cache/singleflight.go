package cache

import (
	"context"
	"strings"

	"golang.org/x/sync/singleflight"

	"github.com/hakhant21/go-starter/internal/metrics"
)

type LoaderFunc[T any] func(ctx context.Context) (T, error)

func GetOrLoad[T any](
	ctx context.Context,
	group *singleflight.Group,
	key string,
	get func(ctx context.Context) (T, bool),
	set func(ctx context.Context, v T),
	load LoaderFunc[T],
) (T, error) {
	prefix := keyPrefix(key)

	if v, ok := get(ctx); ok {
		metrics.CacheHits.WithLabelValues(prefix).Inc()
		return v, nil
	}
	metrics.CacheMisses.WithLabelValues(prefix).Inc()

	v, err, shared := group.Do(key, func() (any, error) {
		if v, ok := get(ctx); ok {
			return v, nil
		}
		metrics.SingleflightLoads.WithLabelValues(prefix).Inc()
		loaded, err := load(ctx)
		if err != nil {
			return nil, err
		}
		set(ctx, loaded)
		return loaded, nil
	})
	if err != nil {
		var zero T
		return zero, err
	}
	if shared {
		metrics.SingleflightShared.WithLabelValues(prefix).Inc()
	}
	return v.(T), nil
}

func keyPrefix(key string) string {
	if i := strings.Index(key, ":"); i > 0 {
		return key[:i]
	}
	return key
}

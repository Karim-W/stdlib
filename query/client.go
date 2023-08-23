package query

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
)

type (
	QueryFn     func(ctx context.Context) (interface{}, error)
	CachingType string
)

const (
	RedisCaching     CachingType = "redis"
	MemoryCaching    CachingType = "memory"
	key_fetch_prefix             = "gqc_fetch_"
)

type Options struct {
	Keys           []string
	CacheTime      time.Duration
	RevalidateTime time.Duration
	Retries        uint
	CacheType      CachingType
}

type QueryClient interface {
	Query(
		ctx context.Context,
		queryFunction QueryFn,
		result interface{},
		options *Options,
	) error
	Mutate(
		ctx context.Context,
		mutationFunction QueryFn,
		result interface{},
		options *Options,
	) error
}

type _QueryClient struct {
	memCache   *cache.Cache
	redisCache *redis.Client
	values     map[string]_QueryResult
	mtx        *sync.RWMutex
}

type _QueryResult struct {
	invalidAt  time.Time
	retryCount uint
	cacheSink  CachingType
}

func NewQueryClient(
	memCache *cache.Cache,
	redisCache *redis.Client,
) QueryClient {
	return &_QueryClient{
		memCache:   memCache,
		redisCache: redisCache,
		values:     make(map[string]_QueryResult),
		mtx:        &sync.RWMutex{},
	}
}

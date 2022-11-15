package stdlib

import (
	"context"
	"fmt"
	"strings"
	"time"

	tracer "github.com/BetaLixT/appInsightsTrace"
	"github.com/go-redis/redis"
	gc "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

const (
	MEMORY_CACHE_TYPE = "Memory"
	REDIS_CACHE_TYPE  = "Redis"
)

var (
	Err_KEY_NOT_FOUND = fmt.Errorf("key not found")
)

type Cache interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	SetWithExpiration(key string, value interface{}, expiration time.Duration) error
	GetCtx(ctx context.Context, key string) (interface{}, error)
	SetCtx(ctx context.Context, key string, value interface{}) error
	SetWithExpirationCtx(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
	DeleteCtx(ctx context.Context, key string) error
	Keys(pattern string) ([]string, error)
	KeysCtx(ctx context.Context, pattern string) ([]string, error)
	WithLogger(l *zap.Logger) Cache
	WithTracer(t *tracer.AppInsightsCore) Cache
	WithName(name string) Cache
}

type cacheImpl struct {
	typ    string
	name   string
	redis  *redis.Client
	mem    *gc.Cache
	tracer *tracer.AppInsightsCore
	lgr    *zap.Logger
}

func InitRedisCache(
	url string,
) (Cache, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	c := &cacheImpl{
		typ:   REDIS_CACHE_TYPE,
		redis: client,
	}
	c.ping()
	return c, nil
}

func InitMemoryCache(
	expiration time.Duration,
	cleanupInterval time.Duration,
) Cache {
	c := &cacheImpl{
		typ: MEMORY_CACHE_TYPE,
		mem: gc.New(expiration, cleanupInterval),
	}
	return c
}

func (c *cacheImpl) WithLogger(l *zap.Logger) Cache {
	c.lgr = l
	return c
}

func (c *cacheImpl) WithName(name string) Cache {
	c.name = name
	return c
}

func (c *cacheImpl) WithTracer(t *tracer.AppInsightsCore) Cache {
	c.tracer = t
	return c
}

func (c *cacheImpl) ping() {
	crashLgr, err := zap.NewProduction()
	if err != nil {
		panic("Could not create logger for redis dependency: " + err.Error())
	}
	go func() {
		for {
			_, err := c.redis.Ping().Result()
			if err != nil {
				crashLgr.Error("Redis ping failed", zap.Error(err))
				if c.tracer != nil {
					c.tracer.TraceException(context.TODO(), err, 0, map[string]string{
						"error":     "Redis ping failed",
						"message":   err.Error(),
						"cacheType": c.typ,
						"time":      time.Now().UTC().Format(time.RFC3339Nano),
					})
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

// ===============================================================================	Core	Funcs	=========================================================================
func (c *cacheImpl) Get(key string) (interface{}, error) {
	return c.GetCtx(context.TODO(), key)
}
func (c *cacheImpl) GetCtx(ctx context.Context, key string) (interface{}, error) {
	now := time.Now()
	var res interface{}
	var err error
	switch c.typ {
	case REDIS_CACHE_TYPE:
		res, err = c.fetchFromRedisCache(ctx, key)
	case MEMORY_CACHE_TYPE:
		res, err = c.fetchFromMemcache(ctx, key)
	}
	end := time.Now()
	if err != nil {
		c.lgr.Error("[Cache] Error fetching from cache", zap.String("key", key), zap.Error(err))
		if c.tracer != nil {
			c.tracer.TraceException(ctx, err, 0, map[string]string{
				"error":     "Error fetching from cache",
				"message":   "Error fetching from cache",
				"cacheType": c.typ,
				"elasped":   end.Sub(now).String(),
			})
			c.tracer.TraceDependency(ctx, "000", c.typ, c.tracer.ServName, "GET", false, now, end, map[string]string{
				"cacheType": c.typ,
				"key":       key,
			})
		}
		return nil, err
	}
	c.lgr.Info("[Cache] Get", zap.String("key", key), zap.String("elasped", end.Sub(now).String()))
	if c.tracer != nil {
		c.tracer.TraceDependency(ctx, "000", c.typ, c.tracer.ServName, "GET", true, now, end, map[string]string{
			"cacheType": c.typ,
			"key":       key,
		})
	}
	return res, nil
}

func (c *cacheImpl) Set(key string, value interface{}) error {
	return c.SetCtx(context.TODO(), key, value)
}
func (c *cacheImpl) SetCtx(ctx context.Context, key string, value interface{}) error {
	now := time.Now()
	var err error
	switch c.typ {
	case REDIS_CACHE_TYPE:
		err = c.setRedisCache(ctx, key, value)
	case MEMORY_CACHE_TYPE:
		err = c.setMemcache(ctx, key, value)
	}
	end := time.Now()
	if err != nil {
		c.lgr.Error("[Cache] Error setting cache", zap.String("key", key), zap.Error(err))
		if c.tracer != nil {
			c.tracer.TraceException(ctx, err, 0, map[string]string{
				"error":     "Error setting cache",
				"message":   "Error setting cache",
				"cacheType": c.typ,
				"elasped":   end.Sub(now).String(),
			})
			c.tracer.TraceDependency(ctx, "000", c.typ, c.tracer.ServName, "SET", false, now, end, map[string]string{
				"cacheType": c.typ,
				"key":       key,
			})
		}
		return err
	}
	c.lgr.Info("[Cache] Set", zap.String("key", key), zap.String("elasped", end.Sub(now).String()))
	if c.tracer != nil {
		c.tracer.TraceDependency(ctx, "000", c.typ, c.tracer.ServName, "SET", true, now, end, map[string]string{
			"cacheType": c.typ,
			"key":       key,
		})
	}
	return nil
}

func (c *cacheImpl) Delete(key string) error {
	return c.DeleteCtx(context.TODO(), key)
}
func (c *cacheImpl) DeleteCtx(ctx context.Context, key string) error {
	now := time.Now()
	var err error
	switch c.typ {
	case REDIS_CACHE_TYPE:
		err = c.deleteFromRedisCache(ctx, key)
	case MEMORY_CACHE_TYPE:
		err = c.deleteFromMemcache(ctx, key)
	}
	end := time.Now()
	if err != nil {
		c.lgr.Error("[Cache] Error deleting cache", zap.String("key", key), zap.Error(err))
		if c.tracer != nil {
			c.tracer.TraceException(ctx, err, 0, map[string]string{
				"error":     "Error deleting cache",
				"message":   "Error deleting cache",
				"cacheType": c.typ,
				"elasped":   end.Sub(now).String(),
			})
			c.tracer.TraceDependency(ctx, "000", c.typ, c.tracer.ServName, "DEL", false, now, end, map[string]string{
				"cacheType": c.typ,
				"key":       key,
			})
		}
		return err
	}
	c.lgr.Info("[Cache] Delete", zap.String("key", key), zap.String("elasped", end.Sub(now).String()))
	if c.tracer != nil {
		c.tracer.TraceDependency(ctx, "000", c.typ, c.tracer.ServName, "DEL", true, now, end, map[string]string{
			"cacheType": c.typ,
			"key":       key,
		})
	}
	return nil
}
func (c *cacheImpl) Keys(pattern string) ([]string, error) {
	return c.KeysCtx(context.TODO(), pattern)
}

func (c *cacheImpl) KeysCtx(ctx context.Context, pattern string) ([]string, error) {
	now := time.Now()
	var res []string
	var err error
	switch c.typ {
	case REDIS_CACHE_TYPE:
		res, err = c.fetchKeysFromRedisCache(ctx, pattern)
	case MEMORY_CACHE_TYPE:
		res, err = c.fetchKeysFromMemcache(ctx, pattern)
	}
	end := time.Now()
	if err != nil {
		c.lgr.Error("[Cache] Error fetching keys from cache", zap.String("pattern", pattern), zap.Error(err))
		if c.tracer != nil {
			c.tracer.TraceException(ctx, err, 0, map[string]string{
				"error":     "Error fetching keys from cache",
				"message":   "Error fetching keys from cache",
				"cacheType": c.typ,
				"elasped":   end.Sub(now).String(),
			})
			c.tracer.TraceDependency(ctx, "000", c.typ, c.tracer.ServName, "KEYS", false, now, end, map[string]string{
				"cacheType": c.typ,
				"pattern":   pattern,
			})
		}
		return nil, err
	}
	c.lgr.Info("[Cache] Keys", zap.String("pattern", pattern), zap.String("elasped", end.Sub(now).String()))
	if c.tracer != nil {
		c.tracer.TraceDependency(ctx, "000", c.typ, c.tracer.ServName, "KEYS", true, now, end, map[string]string{
			"cacheType": c.typ,
			"pattern":   pattern,
		})
	}
	return res, nil
}

func (c *cacheImpl) SetWithExpirationCtx(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	now := time.Now()
	var err error
	switch c.typ {
	case REDIS_CACHE_TYPE:
		err = c.setWithExpiryRedisCache(ctx, key, value, expiration)
	case MEMORY_CACHE_TYPE:
		err = c.setWithExpiryMemcache(ctx, key, value, expiration)
	}
	end := time.Now()
	if err != nil {
		c.lgr.Error("[Cache] Error setting cache", zap.String("key", key), zap.Error(err))
		if c.tracer != nil {
			c.tracer.TraceException(ctx, err, 0, map[string]string{
				"error":     "Error setting cache",
				"message":   "Error setting cache",
				"cacheType": c.typ,
				"elasped":   end.Sub(now).String(),
			})
			c.tracer.TraceDependency(ctx, "000", c.typ, c.tracer.ServName, "SET", false, now, end, map[string]string{
				"cacheType":  c.typ,
				"key":        key,
				"expiration": expiration.String(),
			})
		}
		return err
	}
	c.lgr.Info("[Cache] Set", zap.String("key", key), zap.String("elasped", end.Sub(now).String()))
	if c.tracer != nil {
		c.tracer.TraceDependency(ctx, "000", c.typ, c.tracer.ServName, "SET", true, now, end, map[string]string{
			"cacheType":  c.typ,
			"key":        key,
			"expiration": expiration.String(),
		})
	}
	return nil
}

func (c *cacheImpl) SetWithExpiration(key string, value interface{}, expiration time.Duration) error {
	return c.SetWithExpirationCtx(context.TODO(), key, value, expiration)
}

//===============================================================================	Redis	Cache	=========================================================================

func (c *cacheImpl) fetchFromRedisCache(
	ctx context.Context,
	key string,
) (interface{}, error) {
	return c.redis.Get(key).Result()
}

func (c *cacheImpl) setRedisCache(
	ctx context.Context,
	key string,
	value interface{},
) error {
	return c.redis.Set(key, value, 0).Err()
}

func (c *cacheImpl) deleteFromRedisCache(
	ctx context.Context,
	key string,
) error {
	return c.redis.Del(key).Err()
}

func (c *cacheImpl) fetchKeysFromRedisCache(
	ctx context.Context,
	key string,
) ([]string, error) {
	return c.redis.Keys(key).Result()
}

func (c *cacheImpl) setWithExpiryRedisCache(
	ctx context.Context,
	key string,
	value interface{},
	expiry time.Duration,
) error {
	return c.redis.Set(key, value, expiry).Err()
}

// ===============================================================================	Memory	Cache	=========================================================================
func (c *cacheImpl) fetchFromMemcache(ctx context.Context, key string) (interface{}, error) {
	v, ok := c.mem.Get(key)
	if !ok {
		return nil, Err_KEY_NOT_FOUND
	}
	return v, nil
}

func (c *cacheImpl) setMemcache(ctx context.Context, key string, value interface{}) error {
	c.mem.Set(key, value, gc.DefaultExpiration)
	return nil
}

func (c *cacheImpl) deleteFromMemcache(ctx context.Context, key string) error {
	c.mem.Delete(key)
	return nil
}
func (c *cacheImpl) fetchKeysFromMemcache(ctx context.Context, pattern string) ([]string, error) {
	keys := c.mem.Items()
	var res []string
	for key, _ := range keys {
		if strings.Contains(key, pattern) {
			res = append(res, key)
		}
	}
	return res, nil
}

func (c *cacheImpl) setWithExpiryMemcache(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	c.mem.Set(key, value, expiry)
	return nil
}

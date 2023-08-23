package query

import (
	"context"
	"encoding/json"
	"time"
)

func (c *_QueryClient) Query(
	ctx context.Context,
	queryFunction QueryFn,
	result interface{},
	options *Options,
) error {
	if options == nil {
		return c.doQuery(ctx, queryFunction, result, options)
	}
	if len(options.Keys) == 0 {
		return c.doQuery(ctx, queryFunction, result, options)
	}
	if options.CacheType == RedisCaching && c.redisCache == nil {
		return Err_ProvidedCachingTypeIsNil
	}
	if options.CacheType == MemoryCaching && c.memCache == nil {
		return Err_ProvidedCachingTypeIsNil
	}
	c.mtx.RLock()
	value, found := c.values[options.Keys[0]]
	c.mtx.RUnlock()
	if !found {
		return c.doQuery(ctx, queryFunction, result, options)
	}
	if value.invalidAt.Before(time.Now()) {
		c.evict(ctx, options.CacheType, options.Keys)
		return c.doQuery(ctx, queryFunction, result, options)
	}
	if value.cacheSink == MemoryCaching {
		val, found := c.memCache.Get(key_fetch_prefix + options.Keys[0])
		if !found {
			return c.doQuery(ctx, queryFunction, result, options)
		}
		// set the result to the returned val
		b, ok := val.([]byte)
		if !ok {
			return c.doQuery(ctx, queryFunction, result, options)
		}
		err := json.Unmarshal(b, &result)
		if err != nil {
			return c.doQuery(ctx, queryFunction, result, options)
		}
		return nil
	}
	if value.cacheSink == RedisCaching {
		val, err := c.redisCache.Get(ctx, key_fetch_prefix+options.Keys[0]).Result()
		if err != nil {
			return c.doQuery(ctx, queryFunction, result, options)
		}
		// set the result to the returned val
		err = json.Unmarshal([]byte(val), &result)
		if err != nil {
			return c.doQuery(ctx, queryFunction, result, options)
		}
		return nil
	}
	return c.doQuery(ctx, queryFunction, result, options)
}

func (c *_QueryClient) doQuery(
	ctx context.Context,
	queryFunction QueryFn,
	result interface{},
	options *Options,
) error {
	value, err := queryFunction(ctx)
	if err != nil {
		return err
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return err
	}
	if options == nil {
		return nil
	}
	if len(options.Keys) == 0 {
		return nil
	}
	leastDurtation := options.RevalidateTime
	if options.CacheTime < leastDurtation {
		leastDurtation = options.CacheTime
	}
	qResult := _QueryResult{
		invalidAt:  time.Now().Add(leastDurtation),
		retryCount: 0,
		cacheSink:  options.CacheType,
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for i := 0; i < len(options.Keys); i++ {
		c.values[options.Keys[i]] = qResult
		if options.CacheType == MemoryCaching && leastDurtation > 0 {
			c.memCache.Set(key_fetch_prefix+options.Keys[i], b, leastDurtation)
		}
		if options.CacheType == RedisCaching && leastDurtation > 0 {
			err := c.redisCache.Set(
				ctx,
				key_fetch_prefix+options.Keys[i],
				string(b),
				leastDurtation).
				Err()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

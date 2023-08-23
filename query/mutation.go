package query

import (
	"context"
	"encoding/json"
)

func (c *_QueryClient) Mutate(
	ctx context.Context,
	mutationFunction QueryFn,
	result interface{},
	options *Options,
) error {
	val, err := mutationFunction(ctx)
	if err != nil {
		return err
	}
	if result == nil {
		if options != nil {
			if options.CacheType == RedisCaching && c.redisCache == nil {
				return Err_ProvidedCachingTypeIsNil
			}
			if options.CacheType == MemoryCaching && c.memCache == nil {
				return Err_ProvidedCachingTypeIsNil
			}

			c.evict(ctx, options.CacheType, options.Keys)
			return nil
		}
		return nil
	}
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return err
	}
	if options != nil {
		if options.CacheType == RedisCaching && c.redisCache == nil {
			return Err_ProvidedCachingTypeIsNil
		}
		if options.CacheType == MemoryCaching && c.memCache == nil {
			return Err_ProvidedCachingTypeIsNil
		}
		c.evict(ctx, options.CacheType, options.Keys)
		return nil
	}

	return nil
}

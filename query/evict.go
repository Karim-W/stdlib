package query

import "context"

func (c *_QueryClient) evict(ctx context.Context, typ CachingType, keys []string) {
	if len(keys) == 0 {
		return
	}
	for i := 0; i < len(keys); i++ {
		c.mtx.Lock()
		delete(c.values, keys[i])
		c.mtx.Unlock()
		// delete from caches
		if typ == MemoryCaching {
			c.memCache.Delete(key_fetch_prefix + keys[i])
			continue
		}
		if typ == RedisCaching {
			c.redisCache.Del(ctx, key_fetch_prefix+keys[i])
		}
	}
}

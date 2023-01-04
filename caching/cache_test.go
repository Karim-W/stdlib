package caching

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSetWithExpiry(t *testing.T) {
	rdb := InitMemoryCache(time.Second*10, time.Second*10)
	lgr, err := zap.NewProduction()
	assert.Nil(t, err)
	rdb = rdb.WithLogger(lgr)
	err = rdb.SetWithExpiration("test", "hello", time.Second*10)
	assert.Nil(t, err)
}

func TestSetAndGet(t *testing.T) {
	rdb := InitMemoryCache(time.Second*10, time.Second*10)
	lgr, err := zap.NewProduction()
	assert.Nil(t, err)
	rdb = rdb.WithLogger(lgr)
	err = rdb.SetWithExpiration("test", "hello", time.Second*10)
	assert.Nil(t, err)
	val, err := rdb.Get("test")
	assert.Nil(t, err)
	assert.Equal(t, "hello", val)
}

func TestSuccesfulFetch(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	val, err := cachingLib.Get("test")
	assert.Nil(t, err)
	bol, ok := val.(bool)
	assert.True(t, ok)
	assert.True(t, bol)
}

func TestFailedFetch(t *testing.T) {
	cachingLib := createFailedMockCacher()
	val, err := cachingLib.Get("test")
	assert.NotNil(t, err)
	assert.Nil(t, val)
}

func TestSuccesfulFetchCtx(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	val, err := cachingLib.GetCtx(nil, "test")
	assert.Nil(t, err)
	bol, ok := val.(bool)
	assert.True(t, ok)
	assert.True(t, bol)
}

func TestFailedFetchCtx(t *testing.T) {
	cachingLib := createFailedMockCacher()
	val, err := cachingLib.GetCtx(nil, "test")
	assert.NotNil(t, err)
	assert.Nil(t, val)
}

func TestSuccesfulSet(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	err := cachingLib.Set("test", true)
	assert.Nil(t, err)
}

func TestFailedSet(t *testing.T) {
	cachingLib := createFailedMockCacher()
	err := cachingLib.Set("test", true)
	assert.NotNil(t, err)
}

func TestSuccesfulSetCtx(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	err := cachingLib.SetCtx(nil, "test", true)
	assert.Nil(t, err)
}

func TestFailedSetCtx(t *testing.T) {
	cachingLib := createFailedMockCacher()
	err := cachingLib.SetCtx(nil, "test", true)
	assert.NotNil(t, err)
}

func TestSuccesfulSetWithExpiry(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	err := cachingLib.SetWithExpiration("test", true, time.Second*10)
	assert.Nil(t, err)
}

func TestFailedSetWithExpiry(t *testing.T) {
	cachingLib := createFailedMockCacher()
	err := cachingLib.SetWithExpiration("test", true, time.Second*10)
	assert.NotNil(t, err)
}

func TestSuccesfulSetWithExpiryCtx(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	err := cachingLib.SetWithExpirationCtx(nil, "test", true, time.Second*10)
	assert.Nil(t, err)
}

func TestFailedSetWithExpiryCtx(t *testing.T) {
	cachingLib := createFailedMockCacher()
	err := cachingLib.SetWithExpirationCtx(nil, "test", true, time.Second*10)
	assert.NotNil(t, err)
}

func TestSuccesfulDelete(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	err := cachingLib.Delete("test")
	assert.Nil(t, err)
}

func TestFailedDelete(t *testing.T) {
	cachingLib := createFailedMockCacher()
	err := cachingLib.Delete("test")
	assert.NotNil(t, err)
}

func TestSuccesfulDeleteCtx(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	err := cachingLib.DeleteCtx(nil, "test")
	assert.Nil(t, err)
}

func TestFailedDeleteCtx(t *testing.T) {
	cachingLib := createFailedMockCacher()
	err := cachingLib.DeleteCtx(nil, "test")
	assert.NotNil(t, err)
}

func TestWithLogger(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	cachingLib = cachingLib.WithLogger(nil)
	assert.NotNil(t, cachingLib)
}

func TestWithLoggerNil(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	cachingLib = cachingLib.WithLogger(nil)
	assert.NotNil(t, cachingLib)
}

func TestWithTracer(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	cachingLib = cachingLib.WithTracer(nil)
	assert.NotNil(t, cachingLib)
}

func TestWithTracerNil(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	cachingLib = cachingLib.WithTracer(nil)
	assert.NotNil(t, cachingLib)
}

func TestWithName(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	cachingLib = cachingLib.WithName("test")
	assert.NotNil(t, cachingLib)
}

func TestWithNameNil(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	cachingLib = cachingLib.WithName("")
	assert.NotNil(t, cachingLib)
}

func TestKeys(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	keys, err := cachingLib.Keys("")
	assert.Nil(t, err)
	assert.NotNil(t, keys)
}

func TestKeysCtx(t *testing.T) {
	cachingLib := createSuccessMockCacher()
	keys, err := cachingLib.KeysCtx(nil, "")
	assert.Nil(t, err)
	assert.NotNil(t, keys)
}

func TestKeysFailed(t *testing.T) {
	cachingLib := createFailedMockCacher()
	keys, err := cachingLib.Keys("")
	assert.NotNil(t, err)
	assert.Nil(t, keys)
}

func TestKeysCtxFailed(t *testing.T) {
	cachingLib := createFailedMockCacher()
	keys, err := cachingLib.KeysCtx(nil, "")
	assert.NotNil(t, err)
	assert.Nil(t, keys)
}

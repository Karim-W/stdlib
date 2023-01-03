package stdlib

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

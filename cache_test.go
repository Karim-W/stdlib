package stdlib

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestSetWithExpiry(t *testing.T) {
	rdb, err := InitRedisCache("redis://:@localhost:6379/0")
	if err != nil {
		t.Error(err)
	}
	lgr, err := zap.NewProduction()
	if err != nil {
		t.Error(err)
	}
	rdb = rdb.WithLogger(lgr)
	err = rdb.SetWithExpiration("test", "hello", time.Second*10)
	if err != nil {
		t.Error(err)
	}

}

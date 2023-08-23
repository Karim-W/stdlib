package query

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestRunQuery(t *testing.T) {
	memCache := cache.New(cache.NoExpiration, cache.NoExpiration)
	qc := NewQueryClient(
		memCache,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	var fetchAgain dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "Jane Doe",
				Age:  21,
			}, nil // should be ignored since previous query has cached the result
		},
		&fetchAgain,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, fetchAgain)
}

func TestRunMutate(t *testing.T) {
	memCache := cache.New(cache.NoExpiration, cache.NoExpiration)
	qc := NewQueryClient(
		memCache,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	var fetchAgain dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return nil, nil // should be ignored since previous query has cached the result
		},
		&fetchAgain,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, fetchAgain)
	err = qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return nil, nil
		},
		&fetchAgain,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	// query again to check if the cache is cleared
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "Bob Dylan",
				Age:  40,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, dummy)
	// check that the cache is still valid
	var lastFetch dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return nil, nil // should be ignored since previous query has cached the result
		},
		&lastFetch,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, lastFetch)
}

func BenchmarkQuery(b *testing.B) {
	memCache := cache.New(cache.NoExpiration, cache.NoExpiration)
	qc := NewQueryClient(
		memCache,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	var fetchAgain dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(b, err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = qc.Query(
			context.Background(),
			func(ctx context.Context) (interface{}, error) {
				return nil, nil // should be ignored since previous query has cached the result
			},
			&fetchAgain,
			&Options{
				Keys:           []string{"dummy"},
				CacheTime:      10 * time.Second,
				RevalidateTime: 5 * time.Second,
				Retries:        3,
				CacheType:      MemoryCaching,
			},
		)
	}
}

func TestRunKeylessQuery(t *testing.T) {
	memCache := cache.New(cache.NoExpiration, cache.NoExpiration)
	qc := NewQueryClient(
		memCache,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	var fetchAgain dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "Bob Dylan",
				Age:  40,
			}, nil
		},
		&fetchAgain,
		&Options{
			Keys:           []string{},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, fetchAgain)
}

func TestRunCachelessQuery(t *testing.T) {
	memCache := cache.New(cache.NoExpiration, cache.NoExpiration)
	qc := NewQueryClient(
		memCache,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:      []string{"dummy"},
			CacheType: MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	var fetchAgain dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "Bob Dylan",
				Age:  40,
			}, nil
		},
		&fetchAgain,
		&Options{
			Keys:      []string{"dummy"},
			CacheType: MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, fetchAgain)
}

func TestRunKeylessMutate(t *testing.T) {
	memCache := cache.New(cache.NoExpiration, cache.NoExpiration)
	qc := NewQueryClient(
		memCache,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"Dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	var fetchAgain dummyStruct
	err = qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "Bob Dylan",
				Age:  40,
			}, nil
		},
		&fetchAgain,
		&Options{
			Keys:           []string{},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, fetchAgain)
	var lastFetch dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "Bob Dylan",
				Age:  40,
			}, nil
		},
		&lastFetch,
		&Options{
			Keys:           []string{"Dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, lastFetch)
}

func TestRunQueryRedis(t *testing.T) {
	redisUri := os.Getenv("REDIS_URI")
	if redisUri == "" {
		t.Skip("Skipping test as REDIS_URI is not set")
	}
	opts, err := redis.ParseURL(redisUri)
	if err != nil {
		t.Skip("Skipping test as REDIS_URI is not set")
	}
	redisClient := redis.NewClient(opts)
	qc := NewQueryClient(
		nil,
		redisClient,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      2 * time.Minute,
			RevalidateTime: 5 * time.Minute,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	var fetchAgain dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "Jane Doe",
				Age:  21,
			}, nil // should be ignored since previous query has cached the result
		},
		&fetchAgain,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      2 * time.Minute,
			RevalidateTime: 5 * time.Minute,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, fetchAgain)
}

func TestRunMutateRedis(t *testing.T) {
	redisUri := os.Getenv("REDIS_URI")
	if redisUri == "" {
		t.Skip("Skipping test as REDIS_URI is not set")
	}
	opts, err := redis.ParseURL(redisUri)
	if err != nil {
		t.Skip("Skipping test as REDIS_URI is not set")
	}
	redisClient := redis.NewClient(opts)
	qc := NewQueryClient(
		nil,
		redisClient,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	var fetchAgain dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return nil, nil // should be ignored since previous query has cached the result
		},
		&fetchAgain,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, fetchAgain)
	err = qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return nil, nil
		},
		&fetchAgain,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	// query again to check if the cache is cleared
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "Bob Dylan",
				Age:  40,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, dummy)
	// check that the cache is still valid
	var lastFetch dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return nil, nil // should be ignored since previous query has cached the result
		},
		&lastFetch,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, lastFetch)
}

func TestQueryWithError(t *testing.T) {
	qc := NewQueryClient(
		nil,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("dummy error")
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.NotNil(t, err)
	assert.Equal(t, dummyStruct{}, dummy)
}

func TestOptionlessQuery(t *testing.T) {
	qc := NewQueryClient(
		nil,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		nil,
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
}

func TestOptionlessMutuate(t *testing.T) {
	qc := NewQueryClient(
		nil,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		nil,
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
}

func TestResultLessMutate(t *testing.T) {
	qc := NewQueryClient(
		nil,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return nil, nil
		},
		nil,
		nil,
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{}, dummy)
}

func TestResultLessMutateWithOptions(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)
	qc := NewQueryClient(
		c,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return nil, nil
		},
		nil,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{}, dummy)
}

func TestMutateWithError(t *testing.T) {
	qc := NewQueryClient(
		nil,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var dummy dummyStruct
	err := qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("dummy error")
		},
		&dummy,
		nil,
	)
	assert.NotNil(t, err)
	assert.Equal(t, dummyStruct{}, dummy)
}

func TestActOfGodCacheGone(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)
	qc := NewQueryClient(
		c,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// do a query to populate the cache
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	// delete the cache
	c.Delete(key_fetch_prefix + "dummy")
	// do a query to see if it repopulates
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "bob dylan",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "bob dylan",
		Age:  20,
	}, dummy)
}

func TestQueryWithManipulatedCacheValuesIncorrectType(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)
	qc := NewQueryClient(
		c,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// do a query to populate the cache
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	// manipulate the cache
	c.Set(key_fetch_prefix+"dummy", 12, cache.DefaultExpiration)
	// do a query to see if it falls back to the query function
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "bob dylan",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "bob dylan",
		Age:  20,
	}, dummy)
}

func TestQueryWithManipulatedCacheValuesNonJson(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)
	qc := NewQueryClient(
		c,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// do a query to populate the cache
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	// manipulate the cache
	c.Set(key_fetch_prefix+"dummy", []byte("12"), cache.DefaultExpiration)
	// do a query to see if it falls back to the query function
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "bob dylan",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "bob dylan",
		Age:  20,
	}, dummy)
}

func TestQueryWithNilMemCache(t *testing.T) {
	qc := NewQueryClient(
		nil,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// do a query to populate the cache
	var dummy dummyStruct
	err := qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.NotNil(t, err)
	assert.Equal(t, dummyStruct{}, dummy)
	assert.Equal(t, Err_ProvidedCachingTypeIsNil, err)
}

func TestMutationWithNilMemCache(t *testing.T) {
	qc := NewQueryClient(
		nil,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// do a query to populate the cache
	err := qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		nil,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.NotNil(t, err)
	assert.Equal(t, Err_ProvidedCachingTypeIsNil, err)
}

func TestMutationWithNilMemCacheAgain(t *testing.T) {
	qc := NewQueryClient(
		nil,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// do a query to populate the cache
	var dummy dummyStruct
	err := qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      MemoryCaching,
		},
	)
	assert.NotNil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	assert.Equal(t, Err_ProvidedCachingTypeIsNil, err)
}

func TestMutationWithNilRedis(t *testing.T) {
	qc := NewQueryClient(
		nil,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// do a query to populate the cache
	err := qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		nil,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.NotNil(t, err)
	assert.Equal(t, Err_ProvidedCachingTypeIsNil, err)
}

func TestMutationWithNilRedisAgain(t *testing.T) {
	qc := NewQueryClient(
		nil,
		nil,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// do a query to populate the cache
	var dummy dummyStruct
	err := qc.Mutate(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.NotNil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	assert.Equal(t, Err_ProvidedCachingTypeIsNil, err)
}

func TestActOfGodRedisCacheGone(t *testing.T) {
	redisUrl := os.Getenv("REDIS_URI")
	if redisUrl == "" {
		t.Skip("Skipping test as REDIS_URI is not set")
	}
	opts, err := redis.ParseURL(redisUrl)
	assert.Nil(t, err)
	redisClient := redis.NewClient(opts)
	qc := NewQueryClient(
		nil,
		redisClient,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// do a query to populate the cache
	var dummy dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Minute,
			RevalidateTime: 5 * time.Minute,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	// delete the cache
	err = redisClient.Del(context.Background(), key_fetch_prefix+"dummy").Err()
	assert.Nil(t, err)
	// do a query to populate the cache
	var dummy2 dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "Bob Dylan",
				Age:  40,
			}, nil
		},
		&dummy2,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, dummy2)
	assert.NotEqual(t, dummy, dummy2)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, dummy2)
}

func TestActOfGodRedisCacheManipulated(t *testing.T) {
	redisUrl := os.Getenv("REDIS_URI")
	if redisUrl == "" {
		t.Skip("Skipping test as REDIS_URI is not set")
	}
	opts, err := redis.ParseURL(redisUrl)
	assert.Nil(t, err)
	redisClient := redis.NewClient(opts)
	qc := NewQueryClient(
		nil,
		redisClient,
	)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// do a query to populate the cache
	var dummy dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "John Doe",
				Age:  20,
			}, nil
		},
		&dummy,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Minute,
			RevalidateTime: 5 * time.Minute,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, dummy)
	// delete the cache
	err = redisClient.Set(context.Background(), key_fetch_prefix+"dummy", "12", time.Minute).Err()
	assert.Nil(t, err)
	// do a query to populate the cache
	var dummy2 dummyStruct
	err = qc.Query(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			return dummyStruct{
				Name: "Bob Dylan",
				Age:  40,
			}, nil
		},
		&dummy2,
		&Options{
			Keys:           []string{"dummy"},
			CacheTime:      10 * time.Second,
			RevalidateTime: 5 * time.Second,
			Retries:        3,
			CacheType:      RedisCaching,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, dummy2)
	assert.NotEqual(t, dummy, dummy2)
	assert.Equal(t, dummyStruct{
		Name: "Bob Dylan",
		Age:  40,
	}, dummy2)
}

func TestBareErrorQuery(t *testing.T) {
	qc := NewQueryClient(nil, nil)
	err := qc.Query(context.Background(), func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("test error")
	}, nil, nil)
	assert.NotNil(t, err)
}

func TestBareBadJSONResponseQuery(t *testing.T) {
	qc := NewQueryClient(nil, nil)
	err := qc.Query(context.Background(), func(ctx context.Context) (interface{}, error) {
		return map[string]interface{}{
			"foo": make(chan int),
		}, nil
	}, nil, nil)
	assert.NotNil(t, err)
}

func TestBareBadJSONResponseQuery2(t *testing.T) {
	qc := NewQueryClient(nil, nil)
	var result map[string]interface{}
	err := qc.Query(context.Background(), func(ctx context.Context) (interface{}, error) {
		return `{"name":what?}`, nil
	}, &result, nil)
	assert.NotNil(t, err)
}

func TestNoCacheTypeQuery(t *testing.T) {
	qc := NewQueryClient(nil, nil)
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var result dummyStruct
	err := qc.Query(context.Background(), func(ctx context.Context) (interface{}, error) {
		return dummyStruct{
			Name: "John Doe",
			Age:  20,
		}, nil
	}, &result, &Options{
		Keys:           []string{"dummy"},
		RevalidateTime: 5 * time.Second,
		Retries:        3,
		CacheTime:      10 * time.Second,
	})
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, result)
	err = qc.Query(context.Background(), func(ctx context.Context) (interface{}, error) {
		return dummyStruct{
			Name: "John Doe",
			Age:  20,
		}, nil
	}, &result, &Options{
		Keys:           []string{"dummy"},
		RevalidateTime: 5 * time.Second,
		Retries:        3,
		CacheTime:      10 * time.Second,
	})
	assert.Nil(t, err)
	assert.Equal(t, dummyStruct{
		Name: "John Doe",
		Age:  20,
	}, result)
}

func TestQueryNoRedisSet(t *testing.T) {
	redisUri := os.Getenv("REDIS_URI")
	if redisUri == "" {
		t.Skip("Skipping test as REDIS_URI is not set")
	}
	opts, err := redis.ParseURL(redisUri)
	assert.Nil(t, err)
	redisClient := redis.NewClient(opts)
	qc := NewQueryClient(
		nil,
		redisClient,
	)
	// clost redis connection
	redisClient.Close()
	type dummyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var result dummyStruct
	err = qc.Query(context.Background(), func(ctx context.Context) (interface{}, error) {
		return dummyStruct{
			Name: "John Doe",
			Age:  20,
		}, nil
	}, &result, &Options{
		Keys:           []string{"dummy"},
		RevalidateTime: 5 * time.Second,
		Retries:        3,
		CacheTime:      10 * time.Second,
		CacheType:      RedisCaching,
	})
	assert.NotNil(t, err)
}

func TestBareBadJSONResponseMutation(t *testing.T) {
	qc := NewQueryClient(nil, nil)
	var result map[string]interface{}
	err := qc.Mutate(context.Background(), func(ctx context.Context) (interface{}, error) {
		return map[string]interface{}{
			"foo": make(chan int),
		}, nil
	}, &result, nil)
	assert.NotNil(t, err)
}

func TestBareBadJSONResponseMutation2(t *testing.T) {
	qc := NewQueryClient(nil, nil)
	var result map[string]interface{}
	err := qc.Mutate(context.Background(), func(ctx context.Context) (interface{}, error) {
		return `{"name":what?}`, nil
	}, &result, nil)
	assert.NotNil(t, err)
}

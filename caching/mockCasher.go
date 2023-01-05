package caching

import (
	"context"
	"fmt"
	"time"

	tracer "github.com/BetaLixT/appInsightsTrace"
	"go.uber.org/zap"
)

type failedMockCache struct{}

func createFailedMockCacher() Cache {
	return &failedMockCache{}
}

var (
	err_Sample_Error = fmt.Errorf("failure")
)

func (f *failedMockCache) Get(key string) (interface{}, error) {
	return nil, err_Sample_Error
}

func (f *failedMockCache) Set(key string, value interface{}) error {
	return err_Sample_Error
}

func (f *failedMockCache) WithLogger(l *zap.Logger) Cache {
	return f
}

func (f *failedMockCache) WithTracer(t *tracer.AppInsightsCore) Cache {
	return f
}

func (m *mockCache) SetWithExpiration(key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (m *mockCache) SetWithExpirationCtx(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (f *failedMockCache) SetWithExpiration(key string, value interface{}, expiration time.Duration) error {
	return err_Sample_Error
}

func (f *failedMockCache) SetWithExpirationCtx(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return err_Sample_Error
}

func (f *failedMockCache) GetCtx(ctx context.Context, key string) (interface{}, error) {
	return nil, err_Sample_Error
}

func (f *failedMockCache) SetCtx(ctx context.Context, key string, value interface{}) error {
	return err_Sample_Error
}
func (f *failedMockCache) Delete(key string) error {
	return err_Sample_Error
}
func (f *failedMockCache) DeleteCtx(ctx context.Context, key string) error {
	return err_Sample_Error
}

func (f *failedMockCache) WithName(name string) Cache {
	return f
}

func (f *failedMockCache) Keys(pattern string) ([]string, error) {
	return nil, err_Sample_Error
}

func (f *failedMockCache) KeysCtx(ctx context.Context, pattern string) ([]string, error) {
	return nil, err_Sample_Error
}

func (m *mockCache) WithName(name string) Cache {
	return m
}

type mockCache struct{}

func createSuccessMockCacher() Cache {
	return &mockCache{}
}

func (m *mockCache) Get(key string) (interface{}, error) {
	return true, nil
}

func (m *mockCache) Set(key string, value interface{}) error {
	return nil
}

func (m *mockCache) WithLogger(l *zap.Logger) Cache {
	return m
}

func (m *mockCache) WithTracer(t *tracer.AppInsightsCore) Cache {
	return m
}

func (m *mockCache) GetCtx(ctx context.Context, key string) (interface{}, error) {
	return true, nil
}

func (m *mockCache) SetCtx(ctx context.Context, key string, value interface{}) error {
	return nil
}

func (m *mockCache) Delete(key string) error {
	return nil
}

func (m *mockCache) DeleteCtx(ctx context.Context, key string) error {
	return nil
}

func (m *mockCache) Keys(pattern string) ([]string, error) {
	return []string{}, nil
}

func (m *mockCache) KeysCtx(ctx context.Context, pattern string) ([]string, error) {
	return []string{}, nil
}

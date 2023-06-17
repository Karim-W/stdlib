package httpclient

type RetryPolicy int8

const (
	NO_RETRY            RetryPolicy = iota
	EXPONENTIAL_BACKOFF RetryPolicy = iota
	CONSTANT_BACKOFF    RetryPolicy = iota
)

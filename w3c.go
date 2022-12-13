package stdlib

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	INVALID_VALUE_COUNT     = "invalid number of values"
	VERSION_INVALID_LENGTH  = "invalid version length"
	TRACEID_INVALID_LENGTH  = "invalid traceid length"
	PARENTID_INVALID_LENGTH = "invalid parentid length"
	FLAG_INVALID_LENGTH     = "invalid flag length"
	TRACEID_IS_ZERO         = "error traceid value is zero"
	PARENTID_IS_ZERO        = "error parentid value is zero"
)

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	} else {
		return b, nil
	}
}

func GenerateTraceIdRaw() ([]byte, error) {
	if raw, err := GenerateRandomBytes(16); err != nil {
		return nil, fmt.Errorf("failed to generate bytes %w", err)
	} else {
		return raw, nil
	}
}

func GenerateTraceId() (string, error) {
	if raw, err := GenerateTraceIdRaw(); err != nil {
		return "", fmt.Errorf("failed to generate trace Id %w", err)
	} else {
		return hex.EncodeToString(raw), nil
	}
}

func GenerateParentIdRaw() ([]byte, error) {
	if raw, err := GenerateRandomBytes(8); err != nil {
		return nil, fmt.Errorf("failed to generate butes %w", err)
	} else {
		return raw, nil
	}
}

// Counting on go compiler to inline these plsplspls :)
func GenerateParentId() (string, error) {
	if raw, err := GenerateParentIdRaw(); err != nil {
		return "", fmt.Errorf("failed to generate parent Id %w", err)
	} else {
		return hex.EncodeToString(raw), nil
	}
}

func GenerateNewTraceparent(sampled bool) (string, error) {
	var flag string
	if sampled {
		flag = "01"
	} else {
		flag = "00"
	}
	tid, err := GenerateTraceId()
	if err != nil {
		return "", fmt.Errorf("failed to generate traceId %w", err)
	}
	pid, err := GenerateParentId()
	if err != nil {
		return "", fmt.Errorf("failed to generate parentId %w", err)
	}
	return fmt.Sprintf(
		"00-%s-%s-%s",
		tid,
		pid,
		flag,
	), nil
}

func ParseTraceparentRaw(
	traceparent string,
) ([]byte, []byte, []byte, []byte, error) {
	values := strings.Split(traceparent, "-")
	if len(values) != 4 {
		return nil, nil, nil, nil, fmt.Errorf(INVALID_VALUE_COUNT)
	}

	vers, err := hex.DecodeString(values[0])
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to decode version %w", err)
	}
	if len(vers) != 1 {
		return nil, nil, nil, nil, fmt.Errorf(VERSION_INVALID_LENGTH)
	}

	tid, err := hex.DecodeString(values[1])
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to decode traceId %w", err)
	}
	if len(tid) != 16 {
		return nil, nil, nil, nil, fmt.Errorf(TRACEID_INVALID_LENGTH)
	}

	pid, err := hex.DecodeString(values[2])
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to decode parentId %w", err)
	}
	if len(pid) != 8 {
		return nil, nil, nil, nil, fmt.Errorf(PARENTID_INVALID_LENGTH)
	}

	flg, err := hex.DecodeString(values[3])
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to decode flag %w", err)
	}
	if len(flg) != 1 {
		return nil, nil, nil, nil, fmt.Errorf(FLAG_INVALID_LENGTH)
	}

	return vers, tid, pid, flg, nil
}

func ParseTraceparent(
	traceparent string,
) (string, string, string, string, error) {
	values := strings.Split(traceparent, "-")
	if len(values) != 4 {
		return "", "", "", "", fmt.Errorf(INVALID_VALUE_COUNT)
	}

	if len(values[0]) != 2 {
		return "", "", "", "", fmt.Errorf(VERSION_INVALID_LENGTH)
	}
	if len(values[1]) != 32 {
		return "", "", "", "", fmt.Errorf(TRACEID_INVALID_LENGTH)
	}
	if len(values[2]) != 16 {
		return "", "", "", "", fmt.Errorf(PARENTID_INVALID_LENGTH)
	}
	if len(values[3]) != 2 {
		return "", "", "", "", fmt.Errorf(FLAG_INVALID_LENGTH)
	}

	return values[0], values[1], values[2], values[3], nil
}

func ValidateTraceIdValue(traceId []byte) error {
	for idx := 0; idx < len(traceId); idx++ {
		if traceId[idx] != 0 {
			return nil
		}
	}
	return fmt.Errorf(TRACEID_IS_ZERO)
}

func ValidateParentIdValue(parentId []byte) error {
	for idx := 0; idx < len(parentId); idx++ {
		if parentId[idx] != 0 {
			return nil
		}
	}
	return fmt.Errorf(TRACEID_IS_ZERO)
}

// WILL BE DEPRECATED WHEN TRACING IS MOVED TO A SEPARATE PACAKGE
// THIS IS HERE TO AVOID CIRCULAR DEPENDENCIES
package sqldb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	} else {
		return b, nil
	}
}

func generateParentIdRaw() ([]byte, error) {
	if raw, err := generateRandomBytes(8); err != nil {
		return nil, fmt.Errorf("failed to generate butes %w", err)
	} else {
		return raw, nil
	}
}

// Counting on go compiler to inline these plsplspls :)
func generateParentId() (string, error) {
	if raw, err := generateParentIdRaw(); err != nil {
		return "", fmt.Errorf("failed to generate parent Id %w", err)
	} else {
		return hex.EncodeToString(raw), nil
	}
}

package util

import (
	"fmt"
	"strconv"
)

// Uint64ToString converts uint64 to string
func Uint64ToString(n uint64) string {
	return strconv.FormatUint(n, 10)
}

// StringToUint64 converts string to uint64
func StringToUint64(s string) (uint64, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint64: %w", err)
	}
	return n, nil
}

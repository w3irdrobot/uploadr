package main

import (
	"encoding/hex"
	"path/filepath"
)

func createAndGetDirectory(shasum []byte) string {
	hash := hex.EncodeToString(shasum)
	dirs := make([]string, 0, 4)

	for i := 0; i < 4; i += 1 {
		start := i * 3
		dirs = append(dirs, hash[start:start+3])
	}

	return filepath.Join(dirs...)
}

package coding

import (
	"fmt"
	"os"
	"sync"
)

// ponytail: global mutation lock, per-path locks if write/edit throughput matters.
var mutationMu sync.Mutex

func persistTempOutput(output string) (string, error) {
	file, err := os.CreateTemp("", "nu-bash-*.log")
	if err != nil {
		return "", fmt.Errorf("create temp output: %w", err)
	}
	defer file.Close()
	if _, err := file.WriteString(output); err != nil {
		return "", fmt.Errorf("write temp output %s: %w", file.Name(), err)
	}
	return file.Name(), nil
}

package micro

import (
	"fmt"
	"log/slog"
	"os"
)

func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		slog.Error(fmt.Sprintf("environment variable %s is empty", key))
		os.Exit(1)
	}

	return value
}

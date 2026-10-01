package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Require(keys ...string) error {
	var missing []string
	for _, key := range keys {
		if val := os.Getenv(key); val == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required enviornment variables: %s", strings.Join(missing, ","))
	}
	return nil
}

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}
	return val
}

func GetInt(key string, fallaback int) int {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallaback
	}
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallaback
	}
	return valAsInt
}

func GetBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}
	valAsBool, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return valAsBool
}

func GetList(key string, fallback []string) []string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}
	parts := strings.Split(val, ",")
	items := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	if len(items) == 0 {
		return fallback
	}
	return items
}

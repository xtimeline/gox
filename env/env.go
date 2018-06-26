package env

import (
	"os"

	"github.com/xtimeline/gox/conv"
)

func GetString(key string, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func GetInt64(key string, def int64) (int64, error) {
	val := os.Getenv(key)
	if val == "" {
		return def, nil
	}
	return conv.ParseInt64(val)
}

func GetInt(key string, def int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return def, nil
	}
	return conv.ParseInt(val)
}

func GetBool(key string, def bool) (bool, error) {
	val := os.Getenv(key)
	if val == "" {
		return def, nil
	}
	return conv.ParseBool(val)
}

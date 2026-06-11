package helpers

import (
	"os"
	"strconv"
)

func GetEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

package config

import (
	"os"
	"strconv"
)

var defaultValues = map[string]interface{}{
	"FSVAULT_DATADIR":     "/tmp",
	"FSVAULT_SECRET_KEYS": "",
}

func StringValue(key string) string {
	if defaultValue, ok := defaultValues[key]; ok {
		return getEnvVar(key, defaultValue.(string)).(string)
	}
	return ""
}

// ValueAsInt gets a string value from the env or default
func IntValue(key string) int {

	if defaultValue, ok := defaultValues[key]; ok {
		return getEnvVar(key, defaultValue.(int)).(int)
	}
	return 0
}

// ValueAsBool gets a string value from the env or default
func BoolValue(key string) bool {

	if defaultValue, ok := defaultValues[key]; ok {
		return getEnvVar(key, defaultValue.(bool)).(bool)
	}
	return false
}

// Private methods here
func getEnvVar(key string, fallback interface{}) interface{} {

	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	switch fallback.(type) {
	case string:
		return value
	case bool:
		valueAsBool, err := strconv.ParseBool(value)
		if err != nil {
			return fallback
		}
		return valueAsBool
	case int:
		valueAsInt, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return valueAsInt
	}
	return fallback
}

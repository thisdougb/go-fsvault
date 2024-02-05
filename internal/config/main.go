package config

import (
	"os"
	"strconv"
)

// Config holding struct
type Config struct{}

var defaultValues = map[string]interface{}{
	"FSVAULT_PATH":        "/tmp",
	"FSVAULT_SECRET_KEYS": "",
}

// ValueAsStr gets a string value from the env or default
func (c *Config) ValueAsStr(key string) string {

	if defaultValue, ok := defaultValues[key]; ok {
		return c.getEnvVar(key, defaultValue.(string)).(string)
	}
	return ""
}

// ValueAsInt gets a string value from the env or default
func (c *Config) ValueAsInt(key string) int {

	if defaultValue, ok := defaultValues[key]; ok {
		return c.getEnvVar(key, defaultValue.(int)).(int)
	}
	return 0
}

// ValueAsBool gets a string value from the env or default
func (c *Config) ValueAsBool(key string) bool {

	if defaultValue, ok := defaultValues[key]; ok {
		return c.getEnvVar(key, defaultValue.(bool)).(bool)
	}
	return false
}

// Private methods here
func (c *Config) getEnvVar(key string, fallback interface{}) interface{} {

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

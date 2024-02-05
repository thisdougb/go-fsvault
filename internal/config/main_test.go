//go:build dev
// +build dev

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// adds our test values, when this file is included with go -tags
func init() {
	defaultValues["_TEST_INT_VALUE"] = 10
	defaultValues["_TEST_STR_VALUE"] = "AAA"
	defaultValues["_TEST_BOOL_VALUE"] = false
}

// Test string value type
func TestValueAsStr(t *testing.T) {

	var cfg *Config // dynamic config settings

	// test: unset potential env var, this should return the Str value in defaultValue[map]
	os.Unsetenv("_TEST_STR_VALUE")
	assert.Equal(t, "AAA", cfg.ValueAsStr("_TEST_STR_VALUE"), "no env var set")

	// test: now override the defaultValue[map] using an env var value
	os.Setenv("_TEST_STR_VALUE", "hello")
	assert.Equal(t, "hello", cfg.ValueAsStr("_TEST_STR_VALUE"), "env var set")
	os.Unsetenv("_TEST_STR_VALUE")

	// test: non-existing key
	assert.Equal(t, "", cfg.ValueAsStr("_TEST_STR_NOKEY"), "no key configured")
}

// Test int value type
func TestValueAsInt(t *testing.T) {

	var cfg *Config // dynamic config settings

	// test: unset potential env var, this should return the int value in defaultValue[map]
	os.Unsetenv("_TEST_INT_VALUE")
	assert.Equal(t, 10, cfg.ValueAsInt("_TEST_INT_VALUE"), "no env var set")

	// test: now override the defaultValue[map] using an env var value
	os.Setenv("_TEST_INT_VALUE", "20")
	assert.Equal(t, 20, cfg.ValueAsInt("_TEST_INT_VALUE"), "env var set")
	os.Unsetenv("_TEST_INT_VALUE")

	// test: now we use a non-int env var, which should be ignored
	os.Setenv("_TEST_INT_VALUE", ";")
	assert.Equal(t, 10, cfg.ValueAsInt("_TEST_INT_VALUE"), "env var not int")
	os.Unsetenv("_TEST_INT_VALUE")

	// test: non-existing key
	assert.Equal(t, 0, cfg.ValueAsInt("_TEST_INT_NOKEY"), "no key configured")
}

// Test bool value type
func TestValueAsBool(t *testing.T) {

	var cfg *Config // dynamic config settings

	// test: unset potential env var, this should return the int value in defaultValue[map]
	os.Unsetenv("_TEST_BOOL_VALUE")
	assert.Equal(t, false, cfg.ValueAsBool("_TEST_BOOL_VALUE"), "no env var set")

	// test: now override the defaultValue[map] using an env var value
	os.Setenv("_TEST_BOOL_VALUE", "true")
	assert.Equal(t, true, cfg.ValueAsBool("_TEST_BOOL_VALUE"), "env var set")
	os.Unsetenv("_TEST_BOOL_VALUE")

	// test: now we use a non-int env var, which should be ignored
	os.Setenv("_TEST_BOOL_VALUE", "hello")
	assert.Equal(t, false, cfg.ValueAsBool("_TEST_BOOL_VALUE"), "env var not int")
	os.Unsetenv("_TEST_BOOL_VALUE")

	// test: non-existing key
	assert.Equal(t, false, cfg.ValueAsBool("_TEST_BOOL_NOKEY"), "no key configured")
}

func TestGetEnvVar(t *testing.T) {

	var cfg *Config // dynamic config settings

	// test: when we set a str env var, should should get that value
	os.Setenv("_TEST_STR_NEW", "isset")
	assert.Equal(t, "isset", cfg.getEnvVar("_TEST_STR_NEW", "isset"), "_TEST_STR_NEW")
	os.Unsetenv("_TEST_STR_NEW")

	// test: when no env var exists we should use the fallback value in 2nd arg
	assert.Equal(t, "fallback", cfg.getEnvVar("_TEST_STR_NEW", "fallback"), "_TEST_STR_NEW")

	// test: when we set an int env var, should should get that value
	os.Setenv("_TEST_INT_NEW", "32")
	assert.Equal(t, 32, cfg.getEnvVar("_TEST_INT_NEW", 1), "_TEST_INT_NEW")
	os.Unsetenv("_TEST_INT_NEW")

	// test: when we set an bool env var, should should get that value
	os.Setenv("_TEST_BOOL_NEW", "true")
	assert.Equal(t, true, cfg.getEnvVar("_TEST_BOOL_NEW", false), "_TEST_BOOL_NEW")
	os.Unsetenv("_TEST_BOOL_NEW")

	// test: when we set a non-int env var, should should get the fallback value
	// this is because we don't convert non-int automatically in getEnvVar
	os.Setenv("TEST_UNKNOWN", "2.2")
	assert.Equal(t, 1.1, cfg.getEnvVar("TEST_UNKNOWN", 1.1), "TEST_UNKNOWN")
	os.Unsetenv("TEST_UNKNOWN")
}

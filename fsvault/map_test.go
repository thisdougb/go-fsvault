//go:build dev

package fsvault

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestValue struct {
	Id string `json:"id"`
}

func TestPutMapValue(t *testing.T) {

	var TestCases = []struct {
		description string
		key         string
		value       TestValue
		result      int
	}{
		{
			description: "add one entry",
			key:         "key1",
			value:       TestValue{"test1"},
			result:      1,
		},
		{
			description: "add second entry",
			key:         "key2",
			value:       TestValue{"test2"},
			result:      2,
		},
		{
			description: "add duplicate entry",
			key:         "key1",
			value:       TestValue{"test3"},
			result:      2,
		},
	}

	testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(testRootDir) // clean up

	testMapKey := "/testmap"

	for _, tc := range TestCases {

		PutMapValue(testRootDir, testMapKey, tc.key, tc.value)

		testMap := GetMap[TestValue](testRootDir, testMapKey)

		assert.Equal(t, tc.result, len(testMap), tc.description)
	}
}

func TestGetMapValue(t *testing.T) {

	var TestCases = []struct {
		description string
		key         string
		value       TestValue
	}{
		{
			description: "first entry",
			key:         "key1",
			value:       TestValue{"value1"},
		},
		{
			description: "no entry",
			key:         "keyA",
			value:       TestValue{},
		},
	}

	testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(testRootDir) // clean up

	testMapKey := "/testmap"

	PutMapValue(testRootDir, testMapKey, "key1", TestValue{"value1"})
	PutMapValue(testRootDir, testMapKey, "key2", TestValue{"value2"})
	PutMapValue(testRootDir, testMapKey, "key3", TestValue{"value3"})

	for _, tc := range TestCases {

		value := GetMapValue[TestValue](testRootDir, testMapKey, tc.key)

		assert.Equal(t, tc.value, value, tc.description)
	}
}

func TestGetMapValueWithLock(t *testing.T) {

	var TestCases = []struct {
		description string
		key         string
		value       TestValue
	}{
		{
			description: "first entry",
			key:         "key1",
			value:       TestValue{"value1"},
		},
		{
			description: "no entry",
			key:         "keyA",
			value:       TestValue{},
		},
	}

	testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(testRootDir) // clean up

	testMapKey := "/testmap"

	PutMapValue(testRootDir, testMapKey, "key1", TestValue{"value1"})
	PutMapValue(testRootDir, testMapKey, "key2", TestValue{"value2"})
	PutMapValue(testRootDir, testMapKey, "key3", TestValue{"value3"})

	for _, tc := range TestCases {

		lock, value := GetMapValueWithLock[TestValue](testRootDir, testMapKey, tc.key)
		lock.Unlock()

		assert.Equal(t, tc.value, value, tc.description)
	}
}

func TestDeleteMapValue(t *testing.T) {

	testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(testRootDir) // clean up

	testMapKey := "/testmap"

	PutMapValue(testRootDir, testMapKey, "key1", TestValue{"value1"})
	PutMapValue(testRootDir, testMapKey, "key2", TestValue{"value2"})
	PutMapValue(testRootDir, testMapKey, "key3", TestValue{"value3"})

	DeleteMapValue[TestValue](testRootDir, testMapKey, "key2")

	value := GetMapValue[TestValue](testRootDir, testMapKey, "key1")
	assert.Equal(t, "value1", value.Id, "key1")

	value = GetMapValue[TestValue](testRootDir, testMapKey, "key2")
	assert.Equal(t, "", value.Id, "key2")

	value = GetMapValue[TestValue](testRootDir, testMapKey, "key3")
	assert.Equal(t, "value3", value.Id, "key3")
}

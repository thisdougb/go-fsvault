//go:build dev

package fsvault

import (
	"encoding/json"
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

	defer setupTestDataDir()()

	testMapKey := "/testmap"
	testMap := make(map[string]TestValue)

	for _, tc := range TestCases {

		PutMapValue(testMapKey, tc.key, tc.value)

		data, _ := Get(testMapKey)
		json.Unmarshal(data, &testMap)

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

	defer setupTestDataDir()()
	testMapKey := "/testmap"

	PutMapValue(testMapKey, "key1", TestValue{"value1"})
	PutMapValue(testMapKey, "key2", TestValue{"value2"})
	PutMapValue(testMapKey, "key3", TestValue{"value3"})

	for _, tc := range TestCases {

		value := GetMapValue[TestValue](testMapKey, tc.key)

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

	defer setupTestDataDir()()
	testMapKey := "/testmap"

	PutMapValue(testMapKey, "key1", TestValue{"value1"})
	PutMapValue(testMapKey, "key2", TestValue{"value2"})
	PutMapValue(testMapKey, "key3", TestValue{"value3"})

	for _, tc := range TestCases {

		lock, value := GetMapValueWithLock[TestValue](testMapKey, tc.key)
		lock.Unlock()

		assert.Equal(t, tc.value, value, tc.description)
	}
}

func TestDeleteMapValue(t *testing.T) {

	defer setupTestDataDir()()
	testMapKey := "/testmap"

	PutMapValue(testMapKey, "key1", TestValue{"value1"})
	PutMapValue(testMapKey, "key2", TestValue{"value2"})
	PutMapValue(testMapKey, "key3", TestValue{"value3"})

	DeleteMapValue[TestValue](testMapKey, "key2")

	value := GetMapValue[TestValue](testMapKey, "key1")
	assert.Equal(t, "value1", value.Id, "key1")

	value = GetMapValue[TestValue](testMapKey, "key2")
	assert.Equal(t, "", value.Id, "key2")

	value = GetMapValue[TestValue](testMapKey, "key3")
	assert.Equal(t, "value3", value.Id, "key3")
}

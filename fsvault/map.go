package fsvault

import (
	"encoding/json"
	"log"
	"strings"
)

// GetMapWithLock returns the map with a lock the caller must
// unlock. If the map doesn't exist, the key lock is still
// returned.
func GetMapWithLock[V any](key string) (Unlocker, map[string]V) {

	lock := keylocker.lock(key)
	data := GetMap[V](key)

	return lock, data
}

// GetMap returns the map at key, or an empty map if it doesn't exist.
func GetMap[V any](key string) map[string]V {

	data := make(map[string]V)

	// get map
	dataBytes, err := Get(key)
	if err != nil {
		return data
	}

	// if we have data then...
	if len(dataBytes) > 0 {
		json.Unmarshal(dataBytes, &data)
	}

	return data
}

// GetMapValueWithLock returns the value and a lock on the map, the caller
// must release the lock.
func GetMapValueWithLock[V any](key string, mapkey string) (Unlocker, V) {

	lock := keylocker.lock(key)
	data := GetMapValue[V](mapkey, key)

	return lock, data
}

func GetMapValue[V any](key string, mapkey string) V {

	var value V

	// get map
	dataBytes, err := Get(key)
	if err != nil {
		// if the changelog doesn't exist yet, that's ok, otherwise...
		if !strings.Contains(err.Error(), "no such file or directory") {
			log.Println("datastore.MapGet():", err)
			return value
		}
	}

	data := make(map[string]V)

	// if we have data then...
	if len(dataBytes) > 0 {
		json.Unmarshal(dataBytes, &data)
	}

	// if the entry exists...
	_, ok := data[mapkey]
	if ok {
		return data[mapkey]
	}

	return value
}

func PutMapValue[V any](mapkey string, key string, value V) {

	// get map
	dataBytes, err := Get(mapkey)
	if err != nil {
		// if the changelog doesn't exist yet, that's ok, otherwise...
		if !strings.Contains(err.Error(), "no such file or directory") {
			log.Println("datastore.MapPut():", err)
			return
		}
	}

	mapData := make(map[string]V)

	// if we have data then...
	if len(dataBytes) > 0 {
		json.Unmarshal(dataBytes, &mapData)
	}

	// if the entry already exists...
	_, ok := mapData[key]
	if ok {
		return
	}

	// add new entry
	mapData[key] = value

	dataBytes, _ = json.Marshal(mapData)

	err = Put(mapkey, dataBytes)
	if err != nil {
		log.Println("datastore.MapPut():", err)
	}
}

func DeleteMapValue[V any](mapkey string, key string) {

	// get map
	dataBytes, err := Get(mapkey)
	if err != nil {
		// if the changelog doesn't exist yet, that's ok, otherwise...
		if !strings.Contains(err.Error(), "no such file or directory") {
			return
		}
	}

	mapData := make(map[string]V)

	// if we have no data then...
	if len(dataBytes) == 0 {
		return
	}

	json.Unmarshal(dataBytes, &mapData)

	// if the entry doesn't exist
	_, ok := mapData[key]
	if !ok {
		return
	}

	delete(mapData, key)

	dataBytes, _ = json.Marshal(mapData)

	Put(mapkey, dataBytes)
}

package fsvault

import (
	"encoding/json"
	"log"
	"strings"
)

// GetMapWithLock returns the map with a lock the caller must
// unlock. If the map doesn't exist, the key lock is still
// returned.
func GetMapWithLock[V any](vaultRoot string, vaultKey string) (Unlocker, map[string]V) {

	lock := keylocker.lock(vaultKey)
	data := GetMap[V](vaultRoot, vaultKey)

	return lock, data
}

// GetMap returns the map at key, or an empty map if it doesn't exist.
func GetMap[V any](vaultRoot string, vaultKey string) map[string]V {

	data := make(map[string]V)

	// get map
	dataBytes, err := Get(vaultRoot, vaultKey)
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
func GetMapValueWithLock[V any](vaultRoot string, vaultKey string, mapKey string) (Unlocker, V) {

	lock := keylocker.lock(vaultKey)
	data := GetMapValue[V](vaultRoot, vaultKey, mapKey)

	return lock, data
}

func GetMapValue[V any](vaultRoot string, vaultKey string, mapKey string) V {

	var value V

	// get map
	dataBytes, err := Get(vaultRoot, vaultKey)
	if err != nil {
		// if the changelog doesn't exist yet, that's ok, otherwise...
		if !strings.Contains(err.Error(), "no such file or directory") {
			log.Println("fsvault.datastore.MapGet():", err)
			return value
		}
	}

	data := make(map[string]V)

	// if we have data then...
	if len(dataBytes) > 0 {
		json.Unmarshal(dataBytes, &data)
	}

	// if the entry exists...
	_, ok := data[mapKey]
	if ok {
		return data[mapKey]
	}

	return value
}

// PutMapValue adds value at mapKey, or overwrites value if it exists
func PutMapValue[V any](vaultRoot string, vaultKey string, mapKey string, value V) {

	// get map assuming any prior read call already has a lock
	dataBytes, err := Get(vaultRoot, vaultKey)
	if err != nil {
		// if the changelog doesn't exist yet, that's ok, otherwise...
		if !strings.Contains(err.Error(), "no such file or directory") {
			log.Println("fsvault.datastore.MapPut():", err)
			return
		}
	}

	mapData := make(map[string]V)

	// if we have data then...
	if len(dataBytes) > 0 {
		json.Unmarshal(dataBytes, &mapData)
	}

	// add/overwrite entry
	mapData[mapKey] = value

	dataBytes, _ = json.Marshal(mapData)

	err = Put(vaultRoot, vaultKey, dataBytes)
	if err != nil {
		log.Println("fsvault.datastore.MapPut():", err)
	}
}

func DeleteMapValue[V any](vaultRoot string, vaultKey string, mapKey string) {

	// get map assuming any prior read call already has a lock
	dataBytes, err := Get(vaultRoot, vaultKey)
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
	_, ok := mapData[mapKey]
	if !ok {
		return
	}

	delete(mapData, mapKey)

	dataBytes, _ = json.Marshal(mapData)

	// Put the map back
	Put(vaultRoot, vaultKey, dataBytes)
}

package main

import (
	"log"

	"github.com/thisdougb/go-fsvault/fsvault"
)

func mapUpdate() {

	vaultKey := "/map/email"

	// a type to show generic map usage
	type EmailEntry struct {
		Email string `json:"email"`
	}

	// map values
	mapKey := "myuid"
	email := EmailEntry{"myemail@mydomain.com"}

	// because we want to update, use WithLock to take control of the map
	lock, value := fsvault.GetMapValueWithLock[EmailEntry](vaultKey, mapKey)
	defer lock.Unlock()

	log.Println("mapUpdate(): initial value:", value)

	// we already have a lock on this map, so we don't use WithLock again
	fsvault.PutMapValue[EmailEntry](vaultKey, mapKey, email)

	// read the whole map
	data := fsvault.GetMap[EmailEntry](vaultKey)

	log.Println("mapUpdate(): map data:", data)
}

func valueUpdate() {

	// put some initial data on file
	fsvault.Put("mysecret", []byte("the wind blows from above"))

	// simply read the data back
	data, _ := fsvault.Get("mysecret")
	log.Println("valueUpdate(): got data:", string(data))

	// read the data WithLock when making a change
	lock, data, _ := fsvault.GetWithLock("mysecret")
	defer lock.Unlock()

	// now write a new secret
	fsvault.Put("mysecret", []byte("the wind blows from below"))

	// re-read the value
	data, _ = fsvault.Get("mysecret")

	log.Println("valueUpdate(): got data:", string(data))
}

func main() {

	// a simple key/value update
	valueUpdate()

	// a generic map update
	mapUpdate()
}

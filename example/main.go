package main

import (
	"log"

	"github.com/thisdougb/go-fsvault/fsvault"
)

func main() {

	// put some initial data on file
	fsvault.Put("mysecret", []byte("the wind blows from above"))

	// simply read the data back
	data, _ := fsvault.Get("mysecret")
	log.Println("got data:", string(data))

	// read the data WithLock when making a change
	lock, data, _ := fsvault.GetWithLock("mysecret")
	defer lock.Unlock()

	// now write a new secret
	fsvault.Put("mysecret", []byte("the wind blows from below"))
}

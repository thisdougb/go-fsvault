package main

import (
	"log"

	"github.com/thisdougb/go-fsvault/fsvault"
)

func main() {
	fsvault.Put("testkey2", []byte("test data"))

	lock, data, err := fsvault.GetWithLock("testkey")
	defer lock.Unlock()
	if err != nil {
		log.Println(err)
	}

	log.Println("got data:", string(data))
}

//go:build dev

package fsvault

import (
	"log"
	"os"
)

// SetupTestVaultPath provides unit tests with a temp dir location which
// can be removed by the return function.
func setupTestDataDir() func() {

	// in ths test method run this code at the beginning to setup a tmp dir
	tmpPath, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		panic(err.Error()) // hard stop if there's a problem
	}

	// override the package var
	datadir = tmpPath

	log.Println("datastore.SetupTest(): using tmp fsvault path", tmpPath)

	// in the test method this returned func() is deferred
	return func() {
		log.Println("datastore.SetupTest(): deferred teardown path", tmpPath)
		os.RemoveAll(tmpPath)
	}
}

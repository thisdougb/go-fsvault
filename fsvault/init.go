package fsvault

import (
	"log"
	"strings"

	"github.com/thisdougb/go-fsvault/internal/config"
)

// package level vars
var (
	datadir        string // root data directory
	encryptionKeys []string
)

func init() {

	// get the root data directory, under which all keys wil be stored.
	datadir = config.StringValue("FSVAULT_DATADIR")
	log.Println("fsvault.init(): using datadir", datadir)

	// try to load encryption keys from env vars
	encryptionKeys = getEncryptionKeysFromEnv()
	if len(encryptionKeys) > 0 {
		log.Println("fsvault.init(): encryption enabled.")
	} else {
		log.Println("fsvault.init(): encryption not enabled.")
	}
}

func getEncryptionKeysFromEnv() []string {

	keysEnvVar := config.StringValue("FSVAULT_SECRET_KEYS")
	for _, k := range strings.Split(keysEnvVar, ",") {

		if len(k) == 16 || len(k) == 24 || len(k) == 32 {
			encryptionKeys = append(encryptionKeys, strings.TrimSpace(k))
		} else {
			log.Println("fsvault.init(): invalid secret key length, ignoring", k)
		}
	}

	return encryptionKeys
}

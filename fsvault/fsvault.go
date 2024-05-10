/*
Package fsvault provides a simple crud abstraction over the filesystem as a
datastore.

Data keys should use '/' as a separator, which keeps the implementation
simple by mirroring the underlying filesystem.

Data is any []byte slice.

Encryption of the data, at rest, is enabled by providing a list of encryption
key strings to the NewFSVault() method.
*/
package fsvault

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/thisdougb/go-fsvault/internal/encryption"
	"github.com/thisdougb/go-fsvault/internal/filedata"
)

// The default filesystem permissions, need to be permissive enough if using
// imported fsvault and the cli. Balance access control with encryption use.
var (
	defaultFilePerm      = os.FileMode(0644)
	defaultDirectoryPerm = os.FileMode(0754)
)

var (
	// only one cipher supported to simplify usage
	cipher = "AES-GCM"
)

// KeyExists returns true if data exists at key, and is read/writeable.
func KeyExists(vaultRoot string, vaultKey string) (bool, error) {

	fullPath := filepath.Join(vaultRoot, filepath.Clean(vaultKey))

	info, err := os.Stat(fullPath)
	if err != nil {
		switch err.(type) {
		case *os.PathError:
			return false, err
		default:
			return false, errors.New("Uncaught error: %v" + err.Error())
		}
	}

	if info.Mode().Perm()&0600 == 0600 {
		return true, nil
	} else {
		return false,
			errors.New("Key exists, but file has unusable permissions")
	}
}

// Delete removes the file or directory (if empty) at key.
func Delete(vaultRoot string, vaultKey string) error {

	fullPath := filepath.Join(vaultRoot, filepath.Clean(vaultKey))

	err := os.Remove(fullPath)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return errors.New("key does not exist")
		}
		if strings.Contains(err.Error(), "directory not empty") {
			return errors.New("key is not empty")
		}
		return err
	}
	return nil
}

// List returns an alphabetically sorted list of the object names found a key.
func List(vaultRoot string, vaultKey string) []string {

	keysFound := []string{}

	fullPath := filepath.Join(vaultRoot, filepath.Clean(vaultKey))

	dir, err := os.Open(fullPath)
	if err != nil {
		return keysFound
	}

	files, err := dir.ReadDir(-1)
	if err != nil {
		return keysFound
	}

	for _, f := range files {

		foundKey := filepath.Join(vaultKey, f.Name())
		if f.IsDir() {
			foundKey = foundKey + "/"
		}
		keysFound = append(keysFound, foundKey)
	}

	slices.Sort(keysFound)
	return keysFound
}

// Put writes data to a file at key, overwriting if the file exists.
//
// If encryption keys are present then the primary (first) key is used to
// encrypt the data.
func Put(vaultRoot string, vaultKey string, data []byte) error {

	fullPath := filepath.Join(vaultRoot, filepath.Clean(vaultKey))

	// slight of hand here. we are really only checking if we can't write to
	// this key. we don't care if there's a file there already, or not.
	_, err := KeyExists(vaultRoot, vaultKey)
	if err != nil {
		switch err.(type) {
		case *os.PathError:
			// continue
		default:
			return err
		}
	}

	fd := filedata.FileData{}
	fd.Data = data

	if cipher != "" && len(encryptionKeys) > 0 {

		// if encryption is enabled add the cipher name. currently
		// only one is supported.
		fd.Cipher = cipher

		cipherData, err := encryption.Encrypt(encryptionKeys[0], data)
		if err != nil {
			return err
		}

		fd.Data = cipherData
	}

	fdJSON, _ := json.Marshal(fd)

	// write the file as the last stage, so we reduce the chances of partial
	// dir/file creation
	if err = os.MkdirAll(filepath.Dir(fullPath),
		defaultDirectoryPerm); err != nil {
		return err
	}

	err = os.WriteFile(fullPath, fdJSON, defaultFilePerm)
	if err != nil {
		return err
	}

	return nil
}

// GetWithLock returns a locked mutex with the data, enabling synchronised
// key updates. The caller must Unlock() the lock.
func GetWithLock(vaultRoot string, vaultKey string) (Unlocker, []byte, error) {

	lock := keylocker.lock(vaultKey)
	data, err := Get(vaultRoot, vaultKey)

	return lock, data, err
}

// Get returns the data at key, or an error.
//
// If encryption keys are present, and a non-primary encryption key successfully
// decrypted the data, then the data is re-stored using the primary encryption
// key. See the main documentation for more on encryption key rollover.
func Get(vaultRoot string, vaultKey string) ([]byte, error) {

	fullPath := filepath.Join(vaultRoot, filepath.Clean(vaultKey))

	fd := &filedata.FileData{}

	filecontent, err := os.ReadFile(fullPath)
	if err != nil {
		return fd.Data, err
	}

	err = json.Unmarshal([]byte(filecontent), fd)
	if err != nil {
		return fd.Data, err
	}

	if fd.Cipher != "" {

		// encryptionKey[0] is the most current
		for i, encryptionKey := range encryptionKeys {

			decryptedData, err := encryption.Decrypt(encryptionKey, fd.Data)

			if err != nil {

				// if we tried all the keys, we can't decrypt
				if i == len(encryptionKeys)-1 {
					return fd.Data, err
				}

				// decryption failed, try the next available key
				continue
			}

			fd.Data = make([]byte, len(decryptedData))
			fd.Data = decryptedData

			// if encryptionKey is an old key, refresh data with the latest key
			if i > 0 {
				log.Println("fsvault.Get(): rolling encryption for data at key", vaultKey)

				// remember, Store() takes a key not the fullPath
				err := Put(vaultRoot, vaultKey, fd.Data)
				if err != nil {
					log.Println("fsvault.Get(): failed data refresh at key", vaultKey)
				}
			}
			break
		}
	}
	return fd.Data, nil
}

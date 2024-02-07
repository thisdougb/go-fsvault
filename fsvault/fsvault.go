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
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/thisdougb/go-fsvault/internal/encryption"
	"github.com/thisdougb/go-fsvault/internal/filedata"
	"github.com/thisdougb/go-fsvault/internal/logger"
)

// log is a type alias to avoid exporting Logger
type logAlias = logger.Logger

// FSVault puts and gets data objects in the datastore under `location` on the
// filesystem.
type FSVault struct {
	location       string
	encryptionKeys []string
	cipher         string
	logAlias
}

// The default filesystem permissions, need to be permissive enough if using
// imported fsvault and the cli. Balance access control with encryption use.
var (
	defaultFilePerm      = os.FileMode(0644)
	defaultDirectoryPerm = os.FileMode(0754)
)

/*
NewFSVault returns an FSVault instance ready to use.

The location is the root path on the filesystem, under which all keys will be
stored. Using a root namepsace is generally a nice safety feature, to avoid
polluing the filesystem by accident.

Passing a list of suitable keys (32-char) enables encryption data on-the-fly.
The first key in the list is considered primary, and always used when putting
data.
*/
func NewFSVault(location string, encryptionKeys ...string) *FSVault {

	f := &FSVault{}
	f.location = location
	f.logAlias.TimeCreated = time.Now()
	f.logAlias.LogId = "fsvault"

	// only one cipher mode at the moment
	if len(encryptionKeys) > 0 && encryptionKeys[0] != "" {
		f.encryptionKeys = encryptionKeys
		f.cipher = "AES-GCM"
	}

	return f
}

// FullFilePath returns the filesystem path to the data file for key
func (f *FSVault) FullFilePath(key string) string {
	return filepath.Join(f.location, filepath.Clean(key))
}

// EnableDebug turns on extra debugging messages, perhaps most useful for the
// cli.
func (f *FSVault) EnableDebug() {
	f.logAlias.Debug = true
}

// KeyExists returns true if data exists at key, and is read/writeable.
func (f *FSVault) KeyExists(key string) (bool, error) {

	fullPath := f.FullFilePath(key)

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

/*
Delete removes the file or directory (if empty) at key.
*/
func (f *FSVault) Delete(key string) error {

	loggerName := "fsvault.Delete()"

	fullPath := f.FullFilePath(key)

	err := os.Remove(fullPath)
	if err != nil {
		f.LogDebug(fmt.Sprintf("%s: %s",
			loggerName,
			err.Error()))

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

/*
List returns an alphabetically sorted list of the object names found a key.
*/
func (f *FSVault) List(key string) []string {

	loggerName := "fsvault.List()"

	keysFound := []string{}

	fullPath := f.FullFilePath(key)

	dir, err := os.Open(fullPath)
	if err != nil {
		f.LogDebug(fmt.Sprintf("%s: %s",
			loggerName,
			err.Error()))
		return keysFound
	}

	files, err := dir.ReadDir(-1)
	if err != nil {
		f.LogDebug(fmt.Sprintf("%s: %s",
			loggerName,
			err.Error()))
		return keysFound
	}

	for _, f := range files {

		foundKey := filepath.Join(key, f.Name())
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
func (f *FSVault) Put(key string, data []byte) error {

	loggerName := "fsvault.Put()"

	fullPath := f.FullFilePath(key)

	// slight of hand here. we are really only checking if we can't write to
	// this key. we don't care if there's a file there already, or not.
	_, err := f.KeyExists(fullPath)
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
	fd.Cipher = f.cipher

	if f.cipher != "" && len(f.encryptionKeys) > 0 {

		cipherData, err := encryption.Encrypt(f.encryptionKeys[0], data)
		if err != nil {

			// permanent error, so tell the user
			if err.Error() == "crypto/aes: invalid key size 8" {

				f.LogError(fmt.Sprintf("%s: invalid encryption key length %s",
					loggerName,
					f.encryptionKeys[0]))
			}

			return err
		}

		fd.Data = cipherData

		f.LogDebug(fmt.Sprintf("%s: encrypted data at key %s",
			loggerName,
			key))

	} else {
		f.LogDebug(fmt.Sprintf("%s: encryption not enabled", loggerName))
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

// Get returns the data at key, or an error.
//
// If encryption keys are present, and a non-primary encryption key successfully
// decrypted the data, then the data is re-stored using the primary encryption
// key. See the main documentation for more on encryption key rollover.
func (f *FSVault) Get(key string) ([]byte, error) {

	loggerName := "fsvault.Get()"

	fullPath := f.FullFilePath(key)

	fd := &filedata.FileData{}

	filecontent, err := os.ReadFile(fullPath)
	if err != nil {
		f.LogDebug(
			fmt.Sprintf("%s: error %s", loggerName, err.Error()))

		return fd.Data, err
	}

	err = json.Unmarshal([]byte(filecontent), fd)
	if err != nil {
		f.LogDebug(
			fmt.Sprintf("%s: %s", loggerName, err.Error()))
		return fd.Data, err
	}

	if fd.Cipher != "" {

		// encryptionKey[0] is the most current
		for i, encryptionKey := range f.encryptionKeys {

			decryptedData, err := encryption.Decrypt(encryptionKey, fd.Data)

			if err != nil {

				// permanent error, so tell the user and return
				if err.Error() == "crypto/aes: invalid key size 8" {

					f.LogDebug(
						fmt.Sprintf("%s: invalid encryption key length %s",
							loggerName,
							encryptionKey))

					return fd.Data, err
				}

				f.LogDebug(
					fmt.Sprintf("%s: decrypt error with key %d = %s",
						loggerName,
						i,
						err.Error()))

				// if we tried all the keys, we can't decrypt
				if i == len(f.encryptionKeys)-1 {
					return fd.Data, err
				}

				// decryption failed, assume another encryptionKey may work
				continue
			}

			fd.Data = make([]byte, len(decryptedData))
			fd.Data = decryptedData
			f.LogDebug(
				fmt.Sprintf("%s: decrypted data at key %s",
					loggerName,
					key))

			// if encryptionKey is an old key, refresh data with the latest key
			if i > 0 {
				f.LogDebug(
					fmt.Sprintf(
						"%s: rolling encryption for data at key %s",
						loggerName,
						key))

				// remember, Store() takes a key not the fullPath
				err := f.Put(key, fd.Data)
				if err != nil {
					f.LogDebug(
						fmt.Sprintf("%s: failed data refresh at key %s",
							loggerName,
							key))
				}
			}
			break
		}
	}
	return fd.Data, nil
}

//go:build dev

package fsvault

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the usual strings, but also test we can't break out of location (and
// destroy the filesystem)
func TestFullFilePath(t *testing.T) {

	testCases := []struct {
		description    string
		location       string
		key            string
		expectFullPath string
	}{
		{
			description:    "empty location string",
			location:       "",
			key:            "/some/test/key",
			expectFullPath: "/some/test/key",
		},
		{
			description:    "system root location",
			location:       "/",
			key:            "/some/test/key",
			expectFullPath: "/some/test/key",
		},
		{
			description:    "multiple slashes in location",
			location:       "////",
			key:            "/some/test/key",
			expectFullPath: "/some/test/key",
		},
		{
			description:    "simple trying to break out of location",
			location:       "/data/",
			key:            "/../test/key",
			expectFullPath: "/data/test/key",
		},
		{
			description:    "multple .. trying to break out of location",
			location:       "/data/",
			key:            "/test/../../key",
			expectFullPath: "/data/key",
		},
	}

	for _, tc := range testCases {

		fs := FSVault{}
		fs.location = tc.location
		result := fs.FullFilePath(tc.key)

		assert.Equal(t, tc.expectFullPath, result, tc.description)
	}

}

/*
Test passing in encryption keys as args, or none.
*/
func TestNewFSVault(t *testing.T) {

	var (
		secretKey1 = "mylongsecdddddwwwwdtmylongsecret"
		secretKey2 = "aaojadsnkdakndasnaddddddddddddds"
	)

	location, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(location) // clean up

	f := NewFSVault(location)
	f.EnableDebug()
	assert.Equal(t, 0, len(f.encryptionKeys), "without keys")

	f = NewFSVault(location, secretKey1)
	assert.Equal(t, 1, len(f.encryptionKeys), "with one key")

	f = NewFSVault(location, secretKey1, secretKey2)
	assert.Equal(t, 2, len(f.encryptionKeys), "with two keys")

	f = NewFSVault(location, []string{secretKey1, secretKey2}...)
	assert.Equal(t, 2, len(f.encryptionKeys), "with array of two keys")

	f = NewFSVault(location, []string{""}...)
	assert.Equal(t, 0, len(f.encryptionKeys), "empty string as a key")
}

/*
Test if key (file and directory) exist.
*/
func TestKeyExists(t *testing.T) {

	location, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(location) // clean up

	// create the full path in order to create a file
	key := "test_key"
	tmpfile := filepath.Join(location, key)

	_, err = os.Create(tmpfile)
	if err != nil {
		log.Fatal(err)
	}

	f := NewFSVault(location)
	f.EnableDebug()

	exists, err := f.KeyExists(key)

	assert.Equal(t, true, exists, "tmpfile")
	assert.Equal(t, nil, err, "tmpfile")

}

/*
Test if keys (file and directory) exist, or not.
*/
func TestKeyDoesNotExist(t *testing.T) {

	testCases := []struct {
		key               string
		expectErrorString string
	}{
		{
			key:               "no-such-file",
			expectErrorString: "no such file or directory",
		},
		{
			key:               "/no-such-directory/",
			expectErrorString: "no such file or directory",
		},
	}

	for _, tc := range testCases {

		location, err := os.MkdirTemp("", "thisdougb-fsvault")
		if err != nil {
			assert.Fail(t, err.Error())
		}
		defer os.RemoveAll(location) // clean up

		f := NewFSVault(location)
		f.EnableDebug()

		exists, err := f.KeyExists(tc.key)

		// the actual error includes our random-path tmpdir, so just check
		// for the common part
		result := strings.Contains(err.Error(), tc.expectErrorString)
		assert.Equal(t, true, result, tc.key)

		assert.Equal(t, false, exists, "testfile")
	}

}

func TestPut(t *testing.T) {

	var (
		secretKey1 = "eheheheheheheheheheheheheheheheh"
		secretKey2 = "mylongsecdddddwwwwdtmylongsecret"
	)

	testCases := []struct {
		description string
		secretKeys  []string
		key         string
		data        string
		expectError error
	}{
		{
			description: "with no keys",
			secretKeys:  []string{},
			key:         "/test/unencrypted-data",
			data:        "some test data",
			expectError: nil,
		},
		{
			description: "with one keys",
			secretKeys:  []string{secretKey1},
			key:         "/test/encrypted-data",
			data:        "some test data",
			expectError: nil,
		},
		{
			description: "with multiple keys",
			secretKeys:  []string{secretKey1, secretKey2},
			key:         "/test/encrypted-data-many-keys",
			data:        "some other test data",
			expectError: nil,
		},
	}

	location, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(location) // clean up

	for _, tc := range testCases {

		f := NewFSVault(location, tc.secretKeys...)
		f.EnableDebug()

		err = f.Put(tc.key, []byte(tc.data))
		assert.Equal(t, tc.expectError, err, tc.description)

		data, err := f.Get(tc.key)
		if err != nil {
			assert.Fail(t, err.Error(), tc.description)
		}
		assert.Equal(t, tc.data, string(data), tc.description)
	}
}

func TestInvalidKeyLength(t *testing.T) {

	testCases := []struct {
		description string
		secretKeys  []string
		key         string
		data        string
		expectError error
	}{
		{
			description: "with no keys",
			secretKeys:  []string{"tooshort"},
			key:         "/test/data",
			data:        "some test data",
			expectError: errors.New("crypto/aes: invalid key size 8"),
		},
	}

	location, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(location) // clean up

	for _, tc := range testCases {

		f := NewFSVault(location, tc.secretKeys...)
		f.EnableDebug()

		err = f.Put(tc.key, []byte(tc.data))
		assert.Equal(t, tc.expectError, err, tc.description)
	}
}

func TestKeyRollover(t *testing.T) {

	var (
		secretKey1 = "eheheheheheheheheheheheheheheheh"
		secretKey2 = "mylongsecdddddwwwwdtmylongsecret"
		secretKey3 = "aaojadsnkdakndasnaddddddddddddds"
		secretData = "some super secret data"
		key        = "testfile"
	)

	location, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(location) // clean up

	// store encrypted data using the first key
	f := NewFSVault(location, secretKey1)
	f.EnableDebug()

	err = f.Put(key, []byte(secretData))
	assert.Equal(t, nil, err, "Put() initial content using secretKey1")

	testDescription := "Get() with secretKey1"
	data, err := f.Get(key)
	if err != nil {
		assert.Fail(t, err.Error(), testDescription)
	}
	assert.Equal(t, secretData, string(data), "Get() data with secretKey1")

	// now re-read the data and test the encryption key is rolled
	// read with secretKey1, re-encrypt with secretKey3
	testDescription = "Get() with rolled keys [secretKey3, secretKey2, secretKey1]"
	f = NewFSVault(location, secretKey3, secretKey2, secretKey1)
	f.EnableDebug()
	data, err = f.Get(key)
	if err != nil {
		assert.Fail(t, err.Error(), testDescription)
	}
	assert.Equal(t, secretData, string(data), testDescription)

	// now re-read the data and test key3 specifically encryption
	testDescription = "Get() with secretKey3"
	f = NewFSVault(location, secretKey3)
	f.EnableDebug()
	data, err = f.Get(key)
	if err != nil {
		assert.Fail(t, err.Error(), testDescription)
	}
	assert.Equal(t, secretData, string(data), testDescription)
}

/*
Test list returns file listing if it exists. Uses Put() so run this test after
TestPut().
*/
func TestList(t *testing.T) {

	testCases := []struct {
		description string
		createKeys  []string
		listKey     string
		expectList  []string
	}{
		{
			description: "no found keys",
			createKeys:  []string{},
			listKey:     "/",
			expectList:  []string{},
		},
		{
			description: "simple two keys",
			createKeys:  []string{"key1", "key2"},
			listKey:     "/",
			expectList:  []string{"/key1", "/key2"},
		},
		{
			description: "simple two keys, one subkey",
			createKeys:  []string{"key1", "key2", "sub/key3"},
			listKey:     "/",
			expectList:  []string{"/key1", "/key2", "/sub/"},
		},
		{
			description: "only one subkey",
			createKeys:  []string{"sub/key3"},
			listKey:     "/",
			expectList:  []string{"/sub/"},
		},
	}

	for _, tc := range testCases {

		location, err := os.MkdirTemp("", "thisdougb-fsvault")
		if err != nil {
			assert.Fail(t, err.Error())
		}
		defer os.RemoveAll(location) // clean up

		f := NewFSVault(location)
		f.EnableDebug()

		for _, k := range tc.createKeys {
			f.Put(k, []byte("some data"))
		}

		list := f.List(tc.listKey)
		assert.Equal(t, tc.expectList, list, tc.description)
	}
}

/*
Test delete removes a file, directory, and doesn't fail if the key doesn't exist.
Or, in the case of a directory, if not empty, is not removed.
*/
func TestDelete(t *testing.T) {

	testCases := []struct {
		description   string
		createKeys    []string
		deleteKey     string
		deletedParent string
		expectList    []string
	}{
		{
			description:   "key to delete does not exist",
			createKeys:    []string{},
			deleteKey:     "/key1",
			deletedParent: "/",
			expectList:    []string{},
		},
		{
			description:   "key to delete exists",
			createKeys:    []string{"/key1"},
			deleteKey:     "/key1",
			deletedParent: "/",
			expectList:    []string{},
		},
		{
			description:   "key to delete exists among others",
			createKeys:    []string{"/key1", "/key2", "/sub/key3"},
			deleteKey:     "/key1",
			deletedParent: "/",
			expectList:    []string{"/key2", "/sub/"},
		},
		{
			description:   "delete subkey",
			createKeys:    []string{"/key1", "/key2", "/sub/key3"},
			deleteKey:     "/sub/key3",
			deletedParent: "/sub",
			expectList:    []string{},
		},
		{
			description:   "try to delete populated key",
			createKeys:    []string{"/key1", "/key2", "/sub/key3"},
			deleteKey:     "/sub",
			deletedParent: "/sub",
			expectList:    []string{"/sub/key3"},
		},
	}

	for _, tc := range testCases {

		location, err := os.MkdirTemp("", "thisdougb-fsvault")
		if err != nil {
			assert.Fail(t, err.Error())
		}
		defer os.RemoveAll(location) // clean up

		f := NewFSVault(location)
		f.EnableDebug()

		for _, k := range tc.createKeys {
			f.Put(k, []byte("some data"))
		}

		// I don't think we need to test the return error
		f.Delete(tc.deleteKey)
		assert.Equal(t, tc.expectList, f.List(tc.deletedParent), tc.description)
	}
}

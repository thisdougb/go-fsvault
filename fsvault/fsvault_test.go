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

/*
Test if key (file and directory) exist.
*/
func TestKeyExists(t *testing.T) {

	testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(testRootDir) // clean up

	// create the full path in order to create a file
	vaultKey := "test_key"
	tmpfile := filepath.Join(testRootDir, vaultKey)

	_, err = os.Create(tmpfile)
	if err != nil {
		log.Fatal(err)
	}

	exists, err := KeyExists(testRootDir, vaultKey)

	assert.Equal(t, true, exists, "tmpfile")
	assert.Equal(t, nil, err, "tmpfile")
}

/*
Test if keys (file and directory) exist, or not.
*/
func TestKeyDoesNotExist(t *testing.T) {

	testCases := []struct {
		vaultKey          string
		expectErrorString string
	}{
		{
			vaultKey:          "no-such-file",
			expectErrorString: "no such file or directory",
		},
		{
			vaultKey:          "/no-such-directory/",
			expectErrorString: "no such file or directory",
		},
	}

	for _, tc := range testCases {

		testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
		if err != nil {
			assert.Fail(t, err.Error())
		}
		defer os.RemoveAll(testRootDir) // clean up

		exists, err := KeyExists(testRootDir, tc.vaultKey)

		// the actual error includes our random-path tmpdir, so just check
		// for the common part
		result := strings.Contains(err.Error(), tc.expectErrorString)
		assert.Equal(t, true, result, tc.vaultKey)

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
		vaultKey    string
		data        string
		expectError error
	}{
		{
			description: "with no keys",
			secretKeys:  []string{},
			vaultKey:    "/test/unencrypted-data",
			data:        "some test data",
			expectError: nil,
		},
		{
			description: "with one keys",
			secretKeys:  []string{secretKey1},
			vaultKey:    "/test/encrypted-data",
			data:        "some test data",
			expectError: nil,
		},
		{
			description: "with multiple keys",
			secretKeys:  []string{secretKey1, secretKey2},
			vaultKey:    "/test/encrypted-data-many-keys",
			data:        "some other test data",
			expectError: nil,
		},
	}

	testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(testRootDir) // clean up

	for _, tc := range testCases {

		err = Put(testRootDir, tc.vaultKey, []byte(tc.data))
		assert.Equal(t, tc.expectError, err, tc.description)

		data, err := Get(testRootDir, tc.vaultKey)
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
		vaultKey    string
		data        string
		expectError error
	}{
		{
			description: "with no keys",
			secretKeys:  []string{"tooshort"},
			vaultKey:    "/test/data",
			data:        "some test data",
			expectError: errors.New("crypto/aes: invalid key size 8"),
		},
	}

	testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(testRootDir) // clean up

	for _, tc := range testCases {

		encryptionKeys = tc.secretKeys // package var

		err = Put(testRootDir, tc.vaultKey, []byte(tc.data))
		assert.Equal(t, tc.expectError, err, tc.description)
	}
}

func TestKeyRollover(t *testing.T) {

	var (
		secretKey1 = "eheheheheheheheheheheheheheheheh"
		secretKey2 = "mylongsecdddddwwwwdtmylongsecret"
		secretKey3 = "aaojadsnkdakndasnaddddddddddddds"
		secretData = "some super secret data"
		vaultKey   = "testfile"
	)

	testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer os.RemoveAll(testRootDir) // clean up

	encryptionKeys = []string{secretKey1} // package var

	// store encrypted data using the first key
	err = Put(testRootDir, vaultKey, []byte(secretData))
	assert.Equal(t, nil, err, "Put() initial content using secretKey1")

	testDescription := "Get() with secretKey1"
	data, err := Get(testRootDir, vaultKey)
	if err != nil {
		assert.Fail(t, err.Error(), testDescription)
	}
	assert.Equal(t, secretData, string(data), "Get() data with secretKey1")

	// now re-read the data and test the encryption key is rolled
	// read with secretKey1, re-encrypt with secretKey3
	testDescription = "Get() with rolled keys [secretKey3, secretKey2, secretKey1]"
	encryptionKeys = []string{secretKey3, secretKey2, secretKey1}
	data, err = Get(testRootDir, vaultKey)
	if err != nil {
		assert.Fail(t, err.Error(), testDescription)
	}
	assert.Equal(t, secretData, string(data), testDescription)

	// now re-read the data and test key3 specifically encryption
	testDescription = "Get() with secretKey3"
	encryptionKeys = []string{secretKey3}
	data, err = Get(testRootDir, vaultKey)
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

		testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
		if err != nil {
			assert.Fail(t, err.Error())
		}
		defer os.RemoveAll(testRootDir) // clean up

		for _, k := range tc.createKeys {
			Put(testRootDir, k, []byte("some data"))
		}

		list := List(testRootDir, tc.listKey)
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

		testRootDir, err := os.MkdirTemp("", "thisdougb-fsvault")
		if err != nil {
			assert.Fail(t, err.Error())
		}
		defer os.RemoveAll(testRootDir) // clean up

		for _, k := range tc.createKeys {
			Put(testRootDir, k, []byte("some data"))
		}

		// I don't think we need to test the return error
		Delete(testRootDir, tc.deleteKey)
		assert.Equal(t, tc.expectList, List(testRootDir, tc.deletedParent), tc.description)
	}
}

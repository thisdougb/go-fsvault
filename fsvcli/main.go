package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/thisdougb/go-fsvault/fsvault"
)

/*
fsvcli provides a command line interface to an fsvault datastore.
*/
func main() {

	refreshCmd := flag.NewFlagSet("refresh", flag.ExitOnError)
	refreshRootDir := refreshCmd.String("rootdir", "", "root vault directory")
	refreshKey := refreshCmd.String("key", "", "key to the data")

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getRootDir := getCmd.String("rootdir", "", "root vault directory")
	getKey := getCmd.String("key", "", "key to the data")

	putCmd := flag.NewFlagSet("put", flag.ExitOnError)
	putRootDir := putCmd.String("rootdir", "", "root vault directory")
	putKey := putCmd.String("key", "", "key to the data")
	putData := putCmd.String("data", "", "data to store")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listRootDir := listCmd.String("rootdir", "", "root vault directory")
	listKey := listCmd.String("key", "", "key to the data")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteRootDir := deleteCmd.String("rootdir", "", "root vault directory")
	deleteKey := deleteCmd.String("key", "", "key to the data")

	if len(os.Args) < 2 {
		fmt.Println(`
The fsvcli tool interacts with an FSVault key/value datastore.

Two environment variables control configuration:

    FSVAULT_SECRET_KEYS   a list of encryption keys, see docs for more information

Usage:

    fsvcli <command> [arguments]

The commands are:

    put       put a value into the datastore
    get       get a value from the datastore
    delete    delete a key in the datastore
    list      list keys at a datastore path
    refresh   refresh encryption for a key/value

Use "fsvcli <command> -h" for more information about a command.

Examples:

    $ fsvcli put -rootdir /tmp -key "/user/23/passphrase" -data "Pssst… The green cow has eaten the maple oatmeal"

    $ fsvcli get -rootdir /tmp -key "/user/23/passphrase"       
    Pssst… The green cow has eaten the maple oatmeal

		`)
		os.Exit(1)
	}

	switch os.Args[1] {

	case "refresh":
		refreshCmd.Parse(os.Args[2:])
		err := refreshDataAtKey(*refreshRootDir, *refreshKey)
		if err != nil {
			os.Exit(1)
		}

	case "get":
		getCmd.Parse(os.Args[2:])
		err := getDataAtKey(*getRootDir, *getKey)
		if err != nil {
			os.Exit(1)
		}

	case "put":
		putCmd.Parse(os.Args[2:])
		err := putDataAtKey(*putRootDir, *putKey, *putData)
		if err != nil {
			os.Exit(1)
		}

	case "list":
		listCmd.Parse(os.Args[2:])
		err := listDataAtKey(*listRootDir, *listKey)
		if err != nil {
			os.Exit(1)
		}
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		err := deleteDataAtKey(*deleteRootDir, *deleteKey)
		if err != nil {
			os.Exit(1)
		}
	}
	os.Exit(0)
}

/*
If we read a path then it is automatically re-encrypted with the newer
encryption key.
*/
func refreshDataAtKey(rootDir string, key string) error {

	_, err := fsvault.Get(rootDir, key)
	if err != nil {
		log.Println("refreshDataAtKey():", err)
		return err
	}

	return nil
}

func putDataAtKey(rootDir string, key string, data string) error {

	err := fsvault.Put(rootDir, key, []byte(data))
	if err != nil {
		log.Println("putDataAtKey():", err)
		return err
	}

	return nil
}

func getDataAtKey(rootDir string, key string) error {

	data, err := fsvault.Get(rootDir, key)
	if err != nil {
		log.Println("getDataAtKey():", err)
		return err
	}

	fmt.Printf("%s\n", string(data))

	return nil
}

func deleteDataAtKey(rootDir string, key string) error {

	err := fsvault.Delete(rootDir, key)
	if err != nil {
		fmt.Printf("delete failed because %s\n", err.Error())
	} else {
		fmt.Printf("deleted key %s\n", key)
	}

	return nil
}

func listDataAtKey(rootDir string, key string) error {

	data := fsvault.List(rootDir, key)

	for _, k := range data {
		fmt.Printf("%s\n", k)
	}

	return nil
}

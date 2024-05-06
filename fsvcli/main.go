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
	refreshKey := refreshCmd.String("key", "", "key to the data")

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getKey := getCmd.String("key", "", "key to the data")

	putCmd := flag.NewFlagSet("put", flag.ExitOnError)
	putKey := putCmd.String("key", "", "key to the data")
	putData := putCmd.String("data", "", "data to store")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listKey := listCmd.String("key", "", "key to the data")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteKey := deleteCmd.String("key", "", "key to the data")

	if len(os.Args) < 2 {
		fmt.Println(`
The fsvcli tool interacts with an FSVault key/value datastore.

Two environment variables control configuration:

    FSVAULT_DATADIR          the datastore filesystem path, defaults to /tmp
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

    $ fsvcli put -key "/user/23/passphrase" -data "Pssst… The green cow has eaten the maple oatmeal"

    $ fsvcli get -key "/user/23/passphrase"       
    Pssst… The green cow has eaten the maple oatmeal

		`)
		os.Exit(1)
	}

	switch os.Args[1] {

	case "refresh":
		refreshCmd.Parse(os.Args[2:])
		err := refreshDataAtKey(*refreshKey)
		if err != nil {
			os.Exit(1)
		}

	case "get":
		getCmd.Parse(os.Args[2:])
		err := getDataAtKey(*getKey)
		if err != nil {
			os.Exit(1)
		}

	case "put":
		putCmd.Parse(os.Args[2:])
		err := putDataAtKey(*putKey, *putData)
		if err != nil {
			os.Exit(1)
		}

	case "list":
		listCmd.Parse(os.Args[2:])
		err := listDataAtKey(*listKey)
		if err != nil {
			os.Exit(1)
		}
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		err := deleteDataAtKey(*deleteKey)
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
func refreshDataAtKey(key string) error {

	_, err := fsvault.Get(key)
	if err != nil {
		log.Println("refreshDataAtKey():", err)
		return err
	}

	return nil
}

func putDataAtKey(key string, data string) error {

	err := fsvault.Put(key, []byte(data))
	if err != nil {
		log.Println("putDataAtKey():", err)
		return err
	}

	return nil
}

func getDataAtKey(key string) error {

	data, err := fsvault.Get(key)
	if err != nil {
		log.Println("getDataAtKey():", err)
		return err
	}

	fmt.Printf("%s\n", string(data))

	return nil
}

func deleteDataAtKey(key string) error {

	err := fsvault.Delete(key)
	if err != nil {
		fmt.Printf("delete failed because %s\n", err.Error())
	} else {
		fmt.Printf("deleted key %s\n", key)
	}

	return nil
}

func listDataAtKey(key string) error {

	data := fsvault.List(key)

	for _, k := range data {
		fmt.Printf("%s\n", k)
	}

	return nil
}

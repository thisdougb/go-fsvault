package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/thisdougb/go-fsvault/fsvault"
	"github.com/thisdougb/go-fsvault/internal/config"
)

var (
	cfg            *config.Config
	encryptionKeys []string
	location       string
)

func init() {
	encryptionKeys = getEncryptionKeysFromEnv()
	location = cfg.ValueAsStr("FSVAULT_PATH")
}

/*
fsvcli provides a command line interface to an fsvault datastore.
*/
func main() {

	refreshCmd := flag.NewFlagSet("refresh", flag.ExitOnError)
	refreshKey := refreshCmd.String("key", "", "key to the data")
	refreshDebug := refreshCmd.Bool("debug", false, "enable debug")

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getKey := getCmd.String("key", "", "key to the data")
	getDebug := getCmd.Bool("debug", false, "enable debug")

	putCmd := flag.NewFlagSet("put", flag.ExitOnError)
	putKey := putCmd.String("key", "", "key to the data")
	putData := putCmd.String("data", "", "data to store")
	putDebug := putCmd.Bool("debug", false, "enable debug")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listKey := listCmd.String("key", "", "key to the data")
	listDebug := listCmd.Bool("debug", false, "enable debug")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteKey := deleteCmd.String("key", "", "key to the data")
	deleteDebug := deleteCmd.Bool("debug", false, "enable debug")

	if len(os.Args) < 2 {
		fmt.Println(`
The fsvcli tool interacts with an FSVault key/value datastore.

Two environment variables control conffiguration:

    FSVAULT_PATH          the datastore filesystem path, defaults to /tmp
    FSVAULT_SECRET_KEYS   a list of encryption keys, see docs for more information

Usage:

    fsvcli <command> [arguments] [-debug]

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
		err := refreshDataAtKey(*refreshKey, *refreshDebug)
		if err != nil {
			os.Exit(1)
		}

	case "get":
		getCmd.Parse(os.Args[2:])
		err := getDataAtKey(*getKey, *getDebug)
		if err != nil {
			os.Exit(1)
		}

	case "put":
		putCmd.Parse(os.Args[2:])
		err := putDataAtKey(*putKey, *putData, *putDebug)
		if err != nil {
			os.Exit(1)
		}

	case "list":
		listCmd.Parse(os.Args[2:])
		err := listDataAtKey(*listKey, *listDebug)
		if err != nil {
			os.Exit(1)
		}
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		err := deleteDataAtKey(*deleteKey, *deleteDebug)
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
func refreshDataAtKey(key string, debug bool) error {

	loggerName := "cli.refreshDataAtKey()"

	fs := fsvault.NewFSVault(location, encryptionKeys...)
	if debug {
		fs.EnableDebug()
	}

	_, err := fs.Get(key)
	if err != nil {
		fs.LogDebug(fmt.Sprintf("%s: %s", loggerName, err.Error()))
		return err
	}

	return nil
}

func putDataAtKey(key string, data string, debug bool) error {

	loggerName := "cli.putDataAtKey()"

	fs := fsvault.NewFSVault(location, encryptionKeys...)
	if debug {
		fs.EnableDebug()
	}

	fs.LogDebug(fmt.Sprintf("%s: full path %s", loggerName,
		fs.FullFilePath(key)))

	err := fs.Put(key, []byte(data))
	if err != nil {
		fs.LogDebug(fmt.Sprintf("%s: %s", loggerName, err.Error()))
		return err
	}

	return nil
}

func getDataAtKey(key string, debug bool) error {

	loggerName := "cli.getDataAtKey()"

	fs := fsvault.NewFSVault(location, encryptionKeys...)
	if debug {
		fs.EnableDebug()
	}

	fs.LogDebug(fmt.Sprintf("%s: full path %s", loggerName,
		fs.FullFilePath(key)))

	data, err := fs.Get(key)
	if err != nil {
		fs.LogDebug(fmt.Sprintf("%s: %s", loggerName, err.Error()))
		return err
	}

	fmt.Printf("%s\n", string(data))

	return nil
}

func deleteDataAtKey(key string, debug bool) error {

	loggerName := "cli.deleteDataAtKey()"

	fs := fsvault.NewFSVault(location, encryptionKeys...)
	if debug {
		fs.EnableDebug()
	}

	fs.LogDebug(fmt.Sprintf("%s: full path %s", loggerName,
		fs.FullFilePath(key)))

	err := fs.Delete(key)
	if err != nil {
		fmt.Printf("delete failed because %s\n", err.Error())
	} else {
		fmt.Printf("deleted key %s\n", key)
	}

	return nil
}

func listDataAtKey(key string, debug bool) error {

	loggerName := "cli.listDataAtKey()"

	fs := fsvault.NewFSVault(location, encryptionKeys...)
	if debug {
		fs.EnableDebug()
	}

	fs.LogDebug(fmt.Sprintf("%s: full path %s", loggerName,
		fs.FullFilePath(key)))

	data := fs.List(key)

	for _, k := range data {

		fmt.Printf("%s\n", k)
	}

	return nil
}

func getEncryptionKeysFromEnv() []string {

	keysEnvVar := cfg.ValueAsStr("FSVAULT_SECRET_KEYS")
	for _, k := range strings.Split(keysEnvVar, ",") {
		encryptionKeys = append(encryptionKeys, strings.TrimSpace(k))
	}

	return encryptionKeys
}

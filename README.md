# fsvault

[![release](https://github.com/thisdougb/go-fsvault/actions/workflows/release.yaml/badge.svg)](https://github.com/thisdougb/go-fsvault/actions/workflows/release.yaml)

## Overview

Package and cli tool for storing data on the filesystem, with encryption.

```
go install github.com/thisdougb/go-fsvault/fsvcli
```

## Walkthrough

### Command Line

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


# Go Library

```
import "github.com/thisdougb/go-fsvault/fsvault"

func NewFSVault(location string, encryptionKeys ...string) *FSVault

func (f *FSVault) KeyExists(key string) (bool, error)
func (f *FSVault) Put(key string, data []byte) error
func (f *FSVault) Get(key string) ([]byte, error)
func (f *FSVault) Delete(key string) error
func (f *FSVault) List(key string) []string
```

Simple code usage:

```
keys := getEncryptionKeysFromEnv()

fs := fsvault.NewFSVault("/data/fsvault/", keys...)
fs.Put("/user/23/passphrase", 
    []byte("Pssst… The green cow has eaten the maple oatmeal"))

data, _ := fs.Get("/user/23/passphrase")
```

## Encryption Key Rollover

Rolling encryption keys doesn't need to be difficult. 
Having 'data encrypted at rest' is simple with fsvault, espeically for a live app.

An example using the cli interface, which I think is easier to see what's going on.
Automatic re-encryption happens the same way when importing this package into your app code.

First export the keys and path vars:

```
$ export FSVAULT_SECRET_KEYS="key1-gsecdddddwwwwdtmylongsecret"
$ export FSVAULT_PATH="/data/fsvault/"
```

Now we can store some encrypted data, and read it back:

```
$ fsvault put -key "/user/23/passphrase" -data "Pssst… The green cow has eaten the maple oatmeal"
$ 
$ fsvault get -key "/user/23/passphrase"                                                        
Pssst… The green cow has eaten the maple oatmeal
$
```

When we need to roll the encryption key, simply prepend it to the list of keys.
Now, when that path is read fsvault re-stores the data with the principle encryption key.

Here I've added the `-debug` flag, to show what's happening:

```
$ export FSVAULT_SECRET_KEYS='key2-ensu6fjyivh26fnr5gbaqw3f6go,key1-gsecdddddwwwwdtmylongsecret'
$ fsvault get -key "/user/23/passphrase" -debug                                           
2024/02/05 10:20:14 DEBUG +0.0s [id=fsvault] cli.getDataAtKey(): full path /data/fsvault/user/23/passphrase
2024/02/05 10:20:14 DEBUG +0.0s [id=fsvault] fsvault.Get(): decrypt error with key 0 = cipher: message authentication failed
2024/02/05 10:20:14 DEBUG +0.0s [id=fsvault] fsvault.Get(): decrypted data at key /user/23/passphrase
2024/02/05 10:20:14 DEBUG +0.0s [id=fsvault] fsvault.Get(): rolling encryption for data at key /user/23/passphrase
2024/02/05 10:20:14 DEBUG +0.0s [id=fsvault] fsvault.Put(): encrypted data at key /user/23/passphrase
Pssst… The green cow has eaten the maple oatmeal
$
```

A simple shell for-loop can read all keys to roll the data encryption, forcing re-encryption.
Once all data is refreshed, we can remove the old encryption key:

```
$ export FSVAULT_SECRET_KEYS='key2-ensu6fjyivh26fnr5gbaqw3f6go'                                 
$ fsvault get -key "/user/23/passphrase"
Pssst… The green cow has eaten the maple oatmeal
```


# fsvault

[![release](https://github.com/thisdougb/go-fsvault/actions/workflows/release.yaml/badge.svg)](https://github.com/thisdougb/go-fsvault/actions/workflows/release.yaml)

## Overview

Package and cli tool for storing data on the filesystem, with encryption.

```
go install github.com/thisdougb/go-fsvault/fsvcli
```

### Features

- A simple key/value store on the filesystem
- Encryption of data at rest
- Storing maps of generic types
- Per key locks for synchronised access

## Walkthrough

### Command Line

Two environment variables control conffiguration:

    FSVAULT_DATADIR       the datastore filesystem path, defaults to /tmp
    FSVAULT_SECRET_KEYS   a list of encryption keys, see docs for more information

Usage:

    fsvcli <command> [arguments]

The commands are:

    put       put a value into the datastore
    get       get a value from the datastore
    delete    delete a key in the datastore
    list      list keys at a datastore path
    refresh   refresh encryption for a key/value

Examples:

    $ fsvcli put -key "/user/23/passphrase" -data "Pssst… The green cow has eaten the maple oatmeal"

    $ fsvcli get -key "/user/23/passphrase"       
    Pssst… The green cow has eaten the maple oatmeal

    $ fsvcli put -h
    Usage of put:
      -data string
        	data to store
      -key string
        	key to the data

### Go Library

The example package shows usage.

Write a value at vault key `/user/23/passphrase``:

```
fsvault.Put("/user/23/passphrase", []byte("the wind blows from above"))
```

Read a value at vault key `/user/23/passphrase`:

```
data, _ := fsvault.Get("/user/23/passphrase")
```

Get a map value (int64), including a lock, at map key, defering the lock release:

```
lock, value := fsvault.GetMapValueWithLock[int64](vaultKey, mapKey)
defer lock.Unlock()
```

## Encryption Key Rollover

Rolling encryption keys doesn't need to be difficult. 
Having 'data encrypted at rest' is simple with fsvault, espeically for a live app.

An example using the cli interface, which I think is easier to see what's going on.
Automatic re-encryption happens the same way when importing this package into your app code.

First export the keys and path vars:

```
$ export FSVAULT_SECRET_KEYS="key1-gsecdddddwwwwdtmylongsecret"
$ export FSVAULT_DATADIR="/data/fsvault/"
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
$ fsvault get -key "/user/23/passphrase"
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


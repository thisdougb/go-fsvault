# fsvault

[![release](https://github.com/thisdougb/fsvault/actions/workflows/release.yaml/badge.svg)](https://github.com/thisdougb/fsvault/actions/workflows/release.yaml)

## Overview

Easy to use package for storing data on the filesystem, with encryption.

```
go get github.com/thisdougb/go-fsvault
```

## Usecase

I use fsvault to store app data, encrypted at rest.
Very simple, and much less to go wrong than other data storage options.

I also use it to store unencrypted pre-generated data for serving web content.

With the cli commands I can roll encryption keys regularly, without all the drama.

## Strategy

- Design the API with a similar feel to a database interface, so it feels like a drop-in replacement. Simple crud type operations.
- Support encrypting content with a low barrier to entry, including easy encryption key rollover.

## Walkthrough

The basics:

```
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

In real life, a simple shell for-loop can read all keys to roll the data encryption.
Once all data is refreshed, we can remove the old encryption key:

```
$ export FSVAULT_SECRET_KEYS='key2-ensu6fjyivh26fnr5gbaqw3f6go'                                 
$ fsvault get -key "/user/23/passphrase"
Pssst… The green cow has eaten the maple oatmeal
```

This rollover happens on-the-fly when using the fsvault package.
The cli interface allows Ops people to force a refresh, in a timely manner.


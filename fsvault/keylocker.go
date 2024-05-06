package fsvault

// from: https://stackoverflow.com/questions/40931373/how-to-gc-a-map-of-mutexes-in-go

// Package provides locking per-key.
// For example, you can acquire a lock for a specific user ID and all other requests for that user ID
// will block until that entry is unlocked (effectively your work load will be run serially per-user ID),
// and yet have work for separate user IDs happen concurrently.

import (
	"fmt"
	"sync"
)

type keyLocker struct {
	keymapLock sync.Mutex                   // synchronisation around the keymap
	keymap     map[interface{}]*keymapEntry // keymap holds the individual key locks
}

type keymapEntry struct {
	keymap    *keyLocker  // point back to keyLocker, so we can synchronize removing this mentry when cnt==0
	entryLock sync.Mutex  // entry-specific lock
	cnt       int         // reference count
	key       interface{} // key in keymap, may be string, int, etc.
}

// Unlocker provides an Unlock method to release the lock.
type Unlocker interface {
	Unlock()
}

// New returns an initalized keyLocker.
func newkeyLocker() *keyLocker {
	return &keyLocker{keymap: make(map[interface{}]*keymapEntry)}
}

// lock acquires a lock corresponding to this key.
// This method will never return nil and Unlock() must be called
// to release the lock when done.
func (kl *keyLocker) lock(key interface{}) Unlocker {

	// read or create entry for this key atomically
	kl.keymapLock.Lock()
	entry, ok := kl.keymap[key]
	if !ok {
		entry = &keymapEntry{keymap: kl, key: key}
		kl.keymap[key] = entry
	}
	entry.cnt++ // ref count
	kl.keymapLock.Unlock()

	// acquire lock, will block here until e.cnt==1
	entry.entryLock.Lock()

	return entry
}

// Unlock releases the lock for this entry.
func (entry *keymapEntry) Unlock() {

	kl := entry.keymap

	// lock the keyLocker map
	kl.keymapLock.Lock()

	// decrement and if needed remove entry atomically
	e, ok := kl.keymap[entry.key]
	if !ok { // entry must exist
		kl.keymapLock.Unlock()
		panic(fmt.Errorf("Unlock requested for key=%v but no entry found", entry.key))
	}
	e.cnt--        // ref count
	if e.cnt < 1 { // if it hits zero then we own it and remove from map
		delete(kl.keymap, entry.key)
	}
	kl.keymapLock.Unlock()

	// now that map stuff is handled, we unlock and let
	// anything else waiting on this key through
	entry.entryLock.Unlock()
}

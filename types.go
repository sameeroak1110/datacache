/* *****************************************************************************
Copyright (c) 2022-2030, sameeroak1110 (sameeroak1110@gmail.com)
All rights reserved.
BSD 3-Clause License.

Package     : github.com/sameeroak1110/datacache
Filename    : github.com/sameeroak1110/datacache/types.go
File-type   : golang source code file

Compiler/Runtime: go version go1.14 linux/amd64

Version History
Version     : 1.0
Author      : sameer oak (sameeroak11@gmail.com)
Description :
- Data cache types.
***************************************************************************** */
package datacache

import (
	"sync"
)


type Key interface{}

type Payload struct {
	KeyList []Key           // Key is of type interface{}. cache record may have multiple keys.
	PDataRec interface{}    // this's actual payload-data. should've been created dynamically, i.e., it should be a pointer.
}

type Rec struct {
	KeyList []Key           // Key is of type interface{}. cache record may have multiple keys.
	PDataRec interface{}    // this's actual payload-data. should've been created dynamically, i.e., it should be a pointer. copied from Payload.PDataRec
	isActive bool           // if false, the record is assumed to be deactivated. each record fetch request should be dishonoured if this flag is unset.
	//isDeleted bool        // if true, the record is scheduled for deletion. deleted record is purged at some very low traffic hour. typically, at 0 hrs.
	refcnt uint

	/* record lock: a successful search through the cache returns a locked-record.
	- any transaction on the record is mutually exclusive.
	- it's the caller's prerogative to unlock the locked-record.
	- any further attempt to lock an already locked-record in the same go-routine results in a deadlock. */
	pRecLock *sync.Mutex
	pUnlockRecLock *sync.Mutex  // used specifically during unlocking.
}


// each specific data-cache has a variable of type CacheStore and
// is created (make) in its (specific data-cache implementation) own init function.
//type CacheStore map[interface{}]*DataCacheRec
type CacheStore map[Key]*Rec

// function types for loading the cache and iteration callback.
type LoadFunc func() (bool, []Payload)
type RecHandlerFunc func(interface{}) bool

type DataCache struct {
	// cache store-lock. there're 2 simple rules for store-lock primitives
	// wr store-lock: It's mutually exclusive for any other store-lock.
	// rd store-lock: It's mutually inclusive for any other rd store-lock but exclusive for wr store-lock
	// essentially, wr store-lock is to be taken when there're changes being made to the cache at grand level.
	// for instance, when a record is added to or removed from the cache. equally, when the cache is iterated.
	// and rd store-lock is to be invoked when typically a cache record is to fetched for update or just to fetch
	// record data.
	cacheLock sync.RWMutex       // cache store-lock. rd store lock/unlock and wr store lock/unlock opeeations.
	cache CacheStore             // actual cache store
	loadfn LoadFunc              // function loads the cache during server boot-up.
	reciteratefn RecHandlerFunc  // each record is handled by iterator.
	cnt int                      // number of records in the cache.
	singletonFlag bool           // should be guarded in WR store lock.
}

//var singletonFlag bool       // should be guarded in WR store lock.

/*type Initiator interface {
	Load() (bool, []Payload)
	RecordHandler(*Rec) bool
} */

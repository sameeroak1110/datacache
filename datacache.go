/* ****************************************************************************
Copyright (c) 2022-2030, sameeroak1110 (sameeroak1110@gmail.com)
All rights reserved.
BSD 3-Clause License.

Package     : github.com/sameeroak1110/datacache
Filename    : github.com/sameeroak1110/datacache/datacache.go
File-type   : golang source code file

Compiler/Runtime: go version go1.14 linux/amd64

Version History
Version     : 1.0.0
Author      : sameer oak (sameeroak1110@gmail.com)
Description :
- A rudimentary data-cache implementation.
**************************************************************************** */
package datacache

import (
	"fmt"
	"sync"
	"errors"
	"reflect"
	"runtime/debug"
)

const mutexLocked = 1


/* *****************************************************************************
Description : Takes read store-lock over the cache. pDataCache is assumed to
have been created beforehand. It's a blokcing call. 

Receiver    :
pDataCache *DataCache: DataCache instance.

Arguments   : NA

Return value: NA

Additional note: Method takes read-lock over the cache store referred to by
pDataCache. It's caller's responsibility to read-unlock the cache store.
Read store-lock is mutually exclusive against write store operation and mutually
inclusive for other read store operations.
***************************************************************************** */
func (pDataCache *DataCache) ReadLock() {
	if pDataCache != nil {
		pDataCache.cacheLock.RLock()
	}
}


/* *****************************************************************************
Description : Releases read store-lock over the cache. pDataCache is assumed to
have been created beforehand. It's a non-blocking call.

Receiver    :
pDataCache *DataCache: DataCache instance.

Arguments   : NA

Return value: NA

Additional note: Method releases read lock over the cache store referred to by
pDataCache. An attempt to unlock an already unlocked cache store results into panic.
***************************************************************************** */
func (pDataCache *DataCache) ReadUnlock() {
	if pDataCache != nil {
		pDataCache.cacheLock.RUnlock()
	}
}


/* *****************************************************************************
Description : Takes write store-lock over the cache. pDataCache is assumed to
have been created beforehand. It's a blocking call.

Receiver    :
pDataCache *DataCache: DataCache instance.

Arguments   : NA

Return value: NA

Additional note: Method takes write-lock over the cache store referred to by
pDataCache. It's caller's responsibility to write-unlock the cache store.
Write store-lock is mutually exclusive against any other store-lock.
***************************************************************************** */
func (pDataCache *DataCache) WriteLock() {
	if pDataCache != nil {
		pDataCache.cacheLock.Lock()
	}
}


/* *****************************************************************************
Description : Releases write store-lock over the cache. pDataCache is assumed to
have been created beforehand. It's a non-blocking call.

Receiver    :
pDataCache *DataCache: DataCache instance.

Arguments   : NA

Return value: NA

Additional note: Method releases write lock over the cache store referred to by
pDataCache. An attempt to unlock an already unlocked cache store results into panic.
***************************************************************************** */
func (pDataCache *DataCache) WriteUnlock() {
	if pDataCache != nil {
		pDataCache.cacheLock.Unlock()
	}
}


/* *****************************************************************************
Description : Locks a single datacache record.

Receiver    :
pRec *Rec: A single datacache record.

Implements  : NA

Arguments   : NA

Return value: NA

Additional note: NA
- It's caller's responsibility to unlock the locked datacache record.
***************************************************************************** */
func (pRec *Rec) RecLock() {
	if (pRec != nil) && (pRec.pRecLock != nil) {
		pRec.pRecLock.Lock()
	}
}


/* *****************************************************************************
Description : Unlocks locked datacache record.

Receiver    :
pRec *Rec: A single datacache record.

Implements  : NA

Arguments   : NA

Return value: NA

Additional note: NA
***************************************************************************** */
func (pRec *Rec) RecUnlock() {
	if pRec != nil {
		fmt.Println("Attempting to unlock record.")
		if pRec.pRecLock != nil {
			state := reflect.ValueOf(pRec.pRecLock).Elem().FieldByName("state")
			isLocked := (state.Int() & mutexLocked) == mutexLocked
			msg := "Record is already in unlocked state."
			if isLocked {
				pRec.pRecLock.Unlock()
				msg = "Record has been unlocked."
			}
			fmt.Println(msg)
		}
		pRec = nil
	}
}


// Following 2 functions, one with WR store-lock and other without WR store-lock are going to add a record in the cache.
// For the one without WR store-lock, it's the caller's prerogative to take appropriate lock and release the same once done.
/* *****************************************************************************
Description :
Record pRec is added in the data-cache. It should've been created dynamically.
There's a specific reason why it's to be created dynamically. When the payload needs to be updated, and if it'sn't
a pointer, it needs to be reassigned for the same key. Pointer to the payload makes it easier to update the payload.
Additionally, PDataRec is an empty interface{}. It can hold address of some data but there doesn't exist anything as
address of an empty interface{}.

Receiver    :
pDataCache *DataCache: Datacache instance.

Implements  : NA

Arguments   :
1> keyList []Key: List of keys that refers to the cache record payload.
2> pRec interface{}: Record payload.
3> recExistsErrFlag bool: If true: error is returned in case a record is found for any key.
If false, record is added forcefully.

Return value:
1> int: Number of records in the cache.
2> error: Error string in case of error.

Additional note:
Record pRec is added in the data-cache and is the payload of node in the hash table.
It should've been created dynamically.
This method takes WR store-lock and releases the same while returning.
Therefore, caller shouldn't take any lock before calling this method.
***************************************************************************** */
func (pDataCache *DataCache) AddRec(keyList []Key, pRec interface{}, recExistsErrFlag bool) (int, error) {
	var err error

	if (pDataCache == nil) || (pRec == nil) {
		err = errors.New("NULL datacache or payload.")
		return -1, err
	}

	pDataCache.cacheLock.Lock()
	defer pDataCache.cacheLock.Unlock()

	if recExistsErrFlag {
		for i, _ := range keyList {
			if _, isOK := pDataCache.cache[keyList[i]]; isOK {
				err = errors.New(fmt.Sprintf("Key \"%s\" exists.", keyList[i]))
				return -1, err  // record exists. therefore, record isn't added.
			}
		}
	}

	pDataCacheRec := &Rec {
		PDataRec: pRec,
		KeyList: keyList,
		isActive: true,
	}
	pDataCacheRec.pRecLock = &sync.Mutex{}

	for i, _ := range keyList {
		pDataCache.cache[keyList[i]] = pDataCacheRec
	}

	pDataCache.cnt = pDataCache.cnt + 1

	return pDataCache.cnt, nil
}


/* *****************************************************************************
Description :
Record pRec is added in the data-cache. It should've been created dynamically.
There's a specific reason why it's to be created dynamically. When the payload needs to be updated, and if it'sn't
a pointer, it needs to be reassigned for the same key. Pointer to the payload makes it easier to update the payload.
Additionally, PDataRec is an empty interface{}. It can hold address of some data but there doesn't exist anything as
address of an empty interface{}.

Receiver    :
pDataCache *DataCache: Datacache instance.

Implements  : NA

Arguments   :
1> keyList []Key: List of keys that refers to the cache record payload.
2> pRec interface{}: Record payload.

Return value:
1> int: Number of records in the cache.
2> error: Error string in case of error.

Additional note:
Record pRec is added in the data-cache and is the payload of node in the hash table.
It should've been created dynamically.
This method takes WR store-lock and releases the same while returning.
Therefore, caller shouldn't take any lock before calling this method.
***************************************************************************** */
func (pDataCache *DataCache) ForceAddRec(keyList []Key, pRec interface{}) (int, error) {
	var err error

	if (pDataCache == nil) || (pRec == nil) {
		err = errors.New("NULL datacache or payload.")
		return -1, err
	}

	pDataCache.cacheLock.Lock()
	defer pDataCache.cacheLock.Unlock()

	for i, _ := range keyList {
		if _, isOK := pDataCache.cache[keyList[i]]; isOK {
			delete(pDataCache.cache, i)
		}
	}

	pDataCacheRec := &Rec {
		PDataRec: pRec,
		KeyList: keyList,
		isActive: true,
	}
	pDataCacheRec.pRecLock = &sync.Mutex{}

	for i, _ := range keyList {
		pDataCache.cache[keyList[i]] = pDataCacheRec
	}

	pDataCache.cnt = pDataCache.cnt + 1

	return pDataCache.cnt, nil
}


// Same as AddRec. The only difference is, function returns newly created cache record of type *Rec in the locked state.
func (pDataCache *DataCache) AddAndGetRec(keyList []Key, pRec interface{}, recExistsErrFlag bool) (int, *Rec, error) {
	var err error

	if (pDataCache == nil) || (pRec == nil) {
		err = errors.New("NULL datacache or payload.")
		return -1, nil, err
	}

	pDataCache.cacheLock.Lock()
	defer pDataCache.cacheLock.Unlock()

	if recExistsErrFlag {
		for i, _ := range keyList {
			if _, isOK := pDataCache.cache[keyList[i]]; isOK {
				err = errors.New(fmt.Sprintf("Key \"%s\" exists.", keyList[i]))
				return -1, nil, err  // record exists. therefore, record isn't added.
			}
		}
	}

	pDataCacheRec := &Rec {
		PDataRec: pRec,
		KeyList: keyList,
		isActive: true,
	}
	pDataCacheRec.pRecLock = &sync.Mutex{}

	for i, _ := range keyList {
		pDataCache.cache[keyList[i]] = pDataCacheRec
	}

	pDataCache.cnt = pDataCache.cnt + 1

	prec, _ := pDataCache.cache[keyList[0]]
	prec.pRecLock.Lock()    // record is locked

	return pDataCache.cnt, prec, nil
}


// Same as ForceAddRec. The only difference is, function returns newly created cache record of type *Rec in the locked state.
func (pDataCache *DataCache) ForceAddAndGetRec(keyList []Key, pRec interface{}) (int, *Rec, error) {
	var err error

	if (pDataCache == nil) || (pRec == nil) {
		err = errors.New("NULL datacache or payload.")
		return -1, nil, err
	}

	pDataCache.cacheLock.Lock()
	defer pDataCache.cacheLock.Unlock()

	for i, _ := range keyList {
		if _, isOK := pDataCache.cache[keyList[i]]; isOK {
			delete(pDataCache.cache, i)
		}
	}

	pDataCacheRec := &Rec {
		PDataRec: pRec,
		KeyList: keyList,
		isActive: true,
	}
	pDataCacheRec.pRecLock = &sync.Mutex{}

	for i, _ := range keyList {
		pDataCache.cache[keyList[i]] = pDataCacheRec
	}

	pDataCache.cnt = pDataCache.cnt + 1

	prec, _ := pDataCache.cache[keyList[0]]
	prec.pRecLock.Lock()    // record is locked

	return pDataCache.cnt, prec, nil
}


/* *****************************************************************************
Description :
Maps existing cache record to an additional new key.

Receiver    :
pDataCache *DataCache: Datacache instance.

Implements  : NA

Arguments   :
1> originalKey Key: Existing key.
2> newKey Key: New additional key.

Return value:
1> int: Number of records in the cache.
2> error: Error string in case of error. nil otherwise.

Additional note:
This method takes WR store-lock and releases the same while returning. Therefore, caller shouldn't take
any lock before calling this method.
***************************************************************************** */
func (pDataCache *DataCache) ReAddRec(originalKey Key, newKey Key) (int, error) {
	var err error

	if pDataCache == nil {
		err = errors.New("NULL datacache.")
		return -1, err
	}

	pDataCache.cacheLock.Lock()
	defer pDataCache.cacheLock.Unlock()

	flag := false
	if pRec, isOK := pDataCache.cache[originalKey]; isOK {
		pRec.KeyList = append(pRec.KeyList, newKey)
		pDataCache.cache[newKey] = pRec
		flag = true
	}

	if !flag {
		err = errors.New(fmt.Sprintf("Missing key: %#v.", originalKey))
		return -1, err
	}

	return pDataCache.cnt, nil
}


// Same as ReAddRec. The only difference is function returns newly created cache record of type *Rec in the locked state.
func (pDataCache *DataCache) ReAddAndGetRec(originalKey Key, newKey Key) (int, *Rec, error) {
	var err error

	if pDataCache == nil {
		err = errors.New("NULL datacache.")
		return -1, nil, err
	}

	pDataCache.cacheLock.Lock()
	defer pDataCache.cacheLock.Unlock()

	flag := false
	if pRec, isOK := pDataCache.cache[originalKey]; isOK {
		pRec.KeyList = append(pRec.KeyList, newKey)
		pDataCache.cache[newKey] = pRec
		flag = true
	}

	if !flag {
		err = errors.New(fmt.Sprintf("Missing key \"%s\".", originalKey))
		return -1, nil, err
	}

	prec, _ := pDataCache.cache[newKey]
	prec.pRecLock.Lock()    // record is locked

	return pDataCache.cnt, prec, nil
}


/* *****************************************************************************
Description :
Record pRec is added in the data-cache. It should've been created dynamically.
There's a specific reason why it's to be created dynamically. When the payload needs to be updated, and if it'sn't
a pointer, it needs to be reassigned for the same key. Pointer to the payload makes it easier to update the payload.
Additionally, PDataRec is an empty interface{}. It can hold address of some data but there doesn't exist anything as
address of an empty interface{}.

Receiver    :
pDataCache *DataCache: Datacache instance.

Implements  : NA

Arguments   :
1> keyList []Key: List of keys that refers to the cache record payload.
2> pRec interface{}: Record payload.
3> recExistsErrFlag bool: If true: error is returned in case a record is found for any key.
If false, record is added forcefully.

Return value:
1> int: Number of records in the cache.
2> error: Error string in case of error. nil otherwise.

Additional note:
Record pRec is added in the data-cache and is the payload of node in the hash table.
It should've been created dynamically.
This method doesn't take WR store-lock and thus doesn't release the same. The method assumes caller has invoked
this method in WR store-lock. Thus, the behaviour is unpredictable in case the calling thread calls this method
without WR store-lock.
***************************************************************************** */
func (pDataCache *DataCache) AddRecWOLock(keyList []Key, pRec interface{}, recExistsErrFlag bool) (int, error) {
	var err error

	if (pDataCache == nil) || (pRec == nil) {
		err = errors.New("NULL datacache or payload.")
		return -1, err
	}

	if recExistsErrFlag {
		for i, _ := range keyList {
			if _, isOK := pDataCache.cache[keyList[i]]; isOK {
				err = errors.New(fmt.Sprintf("Key \"%s\" exists.", keyList[i]))  // record exists. therefore, record isn't added.
				return -1, err
			}
		}
	}

	pDataCacheRec := &Rec {
		PDataRec: pRec,
		KeyList: keyList,
		isActive: true,
	}
	pDataCacheRec.pRecLock = &sync.Mutex{}

	for i, _ := range keyList {
		pDataCache.cache[keyList[i]] = pDataCacheRec
	}

	pDataCache.cnt = pDataCache.cnt + 1

	return pDataCache.cnt, nil
}
func (pDataCache *DataCache) ForceAddRecWOLock(keyList []Key, pRec interface{}) (int, error) {
	var err error

	if (pDataCache == nil) || (pRec == nil) {
		err = errors.New("NULL datacache or payload.")
		return -1, err
	}

	for i, _ := range keyList {
		if _, isOK := pDataCache.cache[keyList[i]]; isOK {
			delete(pDataCache.cache, i)
		}
	}

	pDataCacheRec := &Rec {
		PDataRec: pRec,
		KeyList: keyList,
		isActive: true,
	}
	pDataCacheRec.pRecLock = &sync.Mutex{}

	for i, _ := range keyList {
		pDataCache.cache[keyList[i]] = pDataCacheRec
	}

	pDataCache.cnt = pDataCache.cnt + 1

	return pDataCache.cnt, nil
}


// Same as AddRecWOLock(). The only difference is function returns newly created cache record of type *Rec in the locked state.
func (pDataCache *DataCache) AddAndGetRecWOLock(keyList []Key, pRec interface{}, recExistsErrFlag bool) (int, *Rec, error) {
	var err error

	if (pDataCache == nil) || (pRec == nil) {
		err = errors.New("NULL datacache or payload.")
		return -1, nil, err
	}

	if recExistsErrFlag {
		for i, _ := range keyList {
			if _, isOK := pDataCache.cache[keyList[i]]; isOK {
				err = errors.New(fmt.Sprintf("Key \"%s\" exists.", keyList[i]))  // record exists. therefore, record isn't added.
				return -1, nil, err
			}
		}
	}

	pDataCacheRec := &Rec {
		PDataRec: pRec,
		KeyList: keyList,
		isActive: true,
	}
	pDataCacheRec.pRecLock = &sync.Mutex{}

	for i, _ := range keyList {
		pDataCache.cache[keyList[i]] = pDataCacheRec
	}

	pDataCache.cnt = pDataCache.cnt + 1

	prec, _ := pDataCache.cache[keyList[0]]
	prec.pRecLock.Lock()    // record is locked

	return pDataCache.cnt, prec, nil
}
func (pDataCache *DataCache) ForceAddAndGetRecWOLock(keyList []Key, pRec interface{}) (int, *Rec, error) {
	var err error

	if (pDataCache == nil) || (pRec == nil) {
		err = errors.New("NULL datacache or payload.")
		return -1, nil, err
	}

	for i, _ := range keyList {
		if _, isOK := pDataCache.cache[keyList[i]]; isOK {
			delete(pDataCache.cache, i)
		}
	}

	pDataCacheRec := &Rec {
		PDataRec: pRec,
		KeyList: keyList,
		isActive: true,
	}
	pDataCacheRec.pRecLock = &sync.Mutex{}

	for i, _ := range keyList {
		pDataCache.cache[keyList[i]] = pDataCacheRec
	}

	pDataCache.cnt = pDataCache.cnt + 1

	prec, _ := pDataCache.cache[keyList[0]]
	prec.pRecLock.Lock()    // record is locked

	return pDataCache.cnt, prec, nil
}


/* *****************************************************************************
Description :
Maps existing cache record to an additional new key.

Receiver    :
pDataCache *DataCache: Datacache instance.

Implements  : NA

Arguments   :
1> originalKey Key: Existing key.
2> newKey Key: New additional key.

Return value: 
1> int: Number of records in the cache.
2> error: error string in case of error. nil if a new key is associated to the existing cache record.

Additional note:
This method doesn't take WR store-lock and thus doesn't release the same. The method assumes caller has invoked
this method in WR store-lock. Thus, the behaviour is unpredictable in case the calling thread calls this method
without WR store-lock.
***************************************************************************** */
func (pDataCache *DataCache) ReAddRecWOLock(originalKey Key, newKey Key) (int, error) {
	var err error

	if pDataCache == nil {
		err = errors.New("NULL datacache.")
		return -1, err
	}

	flag := false
	if pRec, isOK := pDataCache.cache[originalKey]; isOK {
		pRec.KeyList = append(pRec.KeyList, newKey)
		pDataCache.cache[newKey] = pRec
		flag = true
	}

	if !flag {
		err = errors.New(fmt.Sprintf("Missing key \"%s\".", originalKey))
		return -1, err
	}

	return pDataCache.cnt, nil
}


// Same as ReAddRecWOLock() . The only difference is, function returns newly created cache record of type *Rec in the locked state.
func (pDataCache *DataCache) ReAddAndGetRecWOLock(originalKey Key, newKey Key) (int, *Rec, error) {
	var err error

	if pDataCache == nil {
		err = errors.New("NULL datacache.")
		return -1, nil, err
	}

	flag := false
	if pRec, isOK := pDataCache.cache[originalKey]; isOK {
		pRec.KeyList = append(pRec.KeyList, newKey)
		pDataCache.cache[newKey] = pRec
		flag = true
	}

	if !flag {
		err = errors.New(fmt.Sprintf("Missing key \"%s\".", originalKey))
		return -1, nil, err
	}

	prec, _ := pDataCache.cache[newKey]
	prec.pRecLock.Lock()    // record is locked

	return pDataCache.cnt, prec, nil
}


/* *****************************************************************************
Description :
Disassociates a key.

Receiver    :
pDataCache *DataCache: Datacache instance.

Implements  : NA

Arguments   :
1> key Key: key of cache record to be removed from cache.

Return value:
1> error: Nil or non-nil error

Additional note:
- Method shouldn't be invoked in any - WR or RD - store-lock. It takes WR store-lock and
releases the same.
***************************************************************************** */
func (pDataCache *DataCache) DeleteKey(key Key) error {
	if pDataCache == nil {
		return errors.New("Nil datacache")
	}

	pDataCache.cacheLock.Lock()
	defer func() {
		pDataCache.cacheLock.Unlock()
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()

	delete(pDataCache.cache, key)
	return nil
}


/* *****************************************************************************
Description :
Deletes a record from datacache. invokes delete() built-in.

Receiver    :
pDataCache *DataCache: Datacache instance.

Implements  : NA

Arguments   :
1> key Key: key of cache record to be removed from cache.

Return value:
1> bool: true if record is removed. false otherwise.
2> int: Number of records remained in the cache.

Additional note:
- Method shouldn't be invoked in any - WR or RD - store-lock. It takes WR store-lock and
releases the same.
***************************************************************************** */
func (pDataCache *DataCache) DeleteRec(key Key) (int, error) {
	if pDataCache == nil {
		return -1, errors.New("Nil datacache")
	}

	pDataCache.cacheLock.Lock()
	/* defer func() {
		pDataCache.cacheLock.Unlock()
		if err = recover().(error); err != nil {
			debug.PrintStack()
		}
	}() */

	pRec, isOK := pDataCache.cache[key]
	if !isOK {  // record with key "key" doesn't exist
		pDataCache.cacheLock.Unlock()
		return -1, errors.New("Key doesn't exist.")
	}

	// We're locking this record. This's tricky, however, serves the purpose.
	// Consider this case wherein there're 2 go-routines, one is doing some work on the cache record, maybe updating 
	// the same, and the other one is removing the same record.
	// The former one has locked the record as is returned through GetRec() method, so that no other go
	// routine will be able to access the same record until this one is done with it. So far so good.
	// However, when the 2nd one kicks in, it removes the same cache record. Recall, this same cache record is being
	// accessed in the former go-routine and now is on the verge of panic the moment this go-routine accesses the cache
	// record. And if this go-routine isn't careful enough to handle the panic in the recover of defer, the panic is
	// going to crash the server as this cache record is actually a dangling reference.
	//
	// The trick is to hold the record lock in an auxiliary variable (which is actually going to be a pointer to type
	// sync.Mutex) and block on the Lock() method. Since it's a blocking call, it helps in resolving contention between
	// the former and the latter go-routines, as depicted in the above example.
	// Either of them wins the contention and other one is blocked.
	pTmpRecLock := pRec.pRecLock  // pTmpRecLock just points to pRec.pRecLock

	/* if pTmpRecLock != nil {
		pTmpRecLock.Lock()            // this go-routing waits on the blocking Lock() in case some other go-routine has already been holding this record.
		delete(pDataCache.cache, key) // once gotten, delete succeeds.
		for _, tmpKey := range pRec.KeyList {  // removes the record in entirety, all keys disassociated.
			delete(pDataCache.cache, tmpKey)
		}
		pRec = nil
		pTmpRecLock.Unlock()
		pTmpRecLock = nil  // that's it, done. pRec will never be in use hereon.
	} */

	pTmpRecLock.Lock() // this go-routing waits on the blocking Lock() in case some other go-routine is already holding this record.
	pTmpRecLock.Unlock()
	keylist := pRec.KeyList
	for _, key := range keylist {
		delete(pDataCache.cache, key)
	}
	pTmpRecLock = nil  // that's it, done. pRec will never be in use hereon.
	pRec = nil

	pDataCache.cnt = pDataCache.cnt - 1
	cnt := pDataCache.cnt
	pDataCache.cacheLock.Unlock()

	return cnt, nil
}


/* *****************************************************************************
Description :
Deletes a record from datacache. invokes delete() built-in.

Receiver    :
pDataCache *DataCache: Datacache instance.

Implements  : NA

Arguments   :
1> key Key: key of cache record to be removed from cache.

Return value:
1> bool: true if record is removed. false otherwise.
2> int: Number of records remained in the cache.

Additional note:
- The caller go-routine must invoke this method in WR store-lock. Method doesn't take WR store-lock
and thus doesn't release the same.
- Thus, the behaviour is unpredictable in case the calling go-routine doesn't call this method in
WR store-lock.
***************************************************************************** */
func (pDataCache *DataCache) DeleteRecWOLock(key Key) (bool, int) {
	if pDataCache == nil {
		return false, 0
	}

	/* defer func() {
		if err = recover().(error); err != nil {
			debug.PrintStack()
		}
	}() */

	pRec, isOK := pDataCache.cache[key]
	if !isOK {  // record with key "key" doesn't exist
		return true, 0
	}

	// We're locking this record. This's tricky, however, serves the purpose.
	// Consider this case wherein there're 2 go-routines, one is doing some work on the cache record, maybe updating 
	// the same, and the other one is removing the same record.
	// The former one has locked the record as is returned through GetRec() method, so that no other go
	// routine will be able to access the same record until this one is done with it. So far so good.
	// However, when the 2nd one kicks in, it removes the same cache record. Recall, this same cache record is being
	// accessed in the former go-routine and now is on the verge of panic the moment this go-routine accesses the cache
	// record. And if this go-routine isn't careful enough to handle the panic in the recover of defer, the panic is
	// going to crash the server as this cache record is actually a dangling reference.
	//
	// The trick is to hold the record lock in an auxiliary variable (which is actually going to be a pointer to type
	// sync.Mutex) and block on the Lock() method. Since it's a blocking call, it helps in resolving contention between
	// the former and the latter go-routines, as depicted in the above example.
	// Either of them wins the contention and other one is blocked.
	pTmpRecLock := pRec.pRecLock  // pTmpRecLock just points to pRec.pRecLock

	/* if pTmpRecLock != nil {
		pTmpRecLock.Lock()            // this go-routing waits on the blocking Lock() in case some other go-routine is already holding this record.
		delete(pDataCache.cache, key) // once gotten, delete succeeds.
		for _, tmpKey := range pRec.KeyList {  // removes the record in entirety, all keys disassociated.
			delete(pDataCache.cache, tmpKey)
		}
		pRec = nil
		pTmpRecLock.Unlock()
		pTmpRecLock = nil  // that's it, done. pRec will never be in use hereon.
	} */

	pTmpRecLock.Lock() // this go-routing waits on the blocking Lock() in case some other go-routine is already holding this record.
	keylist := pRec.KeyList
	for _, key := range keylist {
		delete(pDataCache.cache, key)
	}
	pTmpRecLock.Unlock()
	pTmpRecLock = nil  // that's it, done. pRec will never be in use hereon.
	pRec = nil

	pDataCache.cnt = pDataCache.cnt - 1

	return true, pDataCache.cnt
}


// Removes all records from cache.
// Records are removed, not the cache store.
func (pDataCache *DataCache) DeleteCache() bool {
	if pDataCache == nil {
		return false
	}

	pDataCache.cacheLock.Lock()

	for key, prec := range pDataCache.cache {
		pTmpRecLock := prec.pRecLock // pTmpRecLock just points to pRec.pRecLock
		if pTmpRecLock != nil {
			pTmpRecLock.Lock() // this go-routing waits on the blocking Lock() in case some other go-routine is already holding this record.
		}

		delete(pDataCache.cache, key) // that's it, done. prec will never be in use once all its keys are removed from the map.

		if pTmpRecLock != nil {
			pTmpRecLock.Unlock()
		}
		pTmpRecLock = nil
	}

	pDataCache.cacheLock.Unlock()
	return true
}


// Same as DeleteCache() but without write-storelock.
// it's callers prerogative to invoke write-storelock and release the same once done.
func (pDataCache *DataCache) DeleteCacheWOLock() bool {
	if pDataCache == nil {
		return false
	}

	for key, prec := range pDataCache.cache {
		pTmpRecLock := prec.pRecLock // pTmpRecLock just points to pRec.pRecLock
		if pTmpRecLock != nil {
			pTmpRecLock.Lock() // this go-routing waits on the blocking Lock() in case some other go-routine is already holding this record.
		}

		delete(pDataCache.cache, key)  // that's it, done. prec will never be in use once all its keys are removed from the map.

		if pTmpRecLock != nil {
			pTmpRecLock.Unlock()
		}
		pTmpRecLock = nil
	}

	return true
}


/* ****************************************************************************
Description :
Updates state of the datacache record to active or inactive. isActive flag is set to true
to mark it active and to false to mark it disabled.

Receiver    :
pDataCache *DataCache: DataCache instance.

Implements  : NA

Arguments   :
1> key Key: Key to fetch record from datacache pointed to by pDataCache.
2> recState bool: true to mark state active, false to make it inactive.

Return value: true if state is updated. false if record isn't found in the datacache
pointed to by pDataCache.

Additional note:
- Calling go-routine shouldn't take any lock over the store. It's a deadlock in case it
does so.
**************************************************************************** */
func (pDataCache *DataCache) UpdateRecState(key Key, recState bool) bool {
	if pDataCache == nil {
		return false
	}

	pDataCache.cacheLock.RLock()
	defer pDataCache.cacheLock.RUnlock()

	pRec, isOK := pDataCache.cache[key]
	if !isOK {
		return false
	}

	pRec.pRecLock.Lock()
	pRec.isActive = recState
	pRec.pRecLock.Unlock()

	return true
}


/* ****************************************************************************
Description :
Updates state of the datacache record to active or inactive. isActive flag is set to true
to mark it active and to false to mark it disabled.

Receiver    :
pDataCache *DataCache: DataCache instance.

Implements  : NA

Arguments   :
1> key Key: Key to fetch record from datacache pointed to by pDataCache.
2> recState bool: true to mark state active, false to make it inactive.

Return value: true if state is updated. false if record isn't found in the datacache
pointed to by pDataCache.

Additional note:
- The caller go-routine must invoke this method in WR or RD store-lock. Method doesn't take any store
lock and thus doesn't release the same.
- Thus, the behaviour is unpredictable in case the caller go-routine doesn't call this method in
any store-lock.
- There's one more additional care one must observe. The cache record which is referenced by "key" and
who's state is possibly updated, shouldn't've been in locked state. It's a deadlock otherwise.
**************************************************************************** */
func (pDataCache *DataCache) UpdateRecStateWOLock(key Key, recState bool) bool {
	if pDataCache == nil {
		return false
	}

	pRec, isOK := pDataCache.cache[key]
	if !isOK {
		return false
	}

	pRec.pRecLock.Lock()  // pRec shouldn't've been in locked state. it's a deadlock otherwise.
	pRec.isActive = recState
	pRec.pRecLock.Unlock()

	return true
}


/* ****************************************************************************
Description :
Function returns a locked datacache record. It's caller's responsibility to unlock the
same once done with it. Returned datacache record is a pointer to Rec.

Receiver    :
pDataCache *DataCache: Instance of datacache

Implements  : NA

Arguments   :
1> key Key: Key to the cache record.

Return value:
1> bool: true if successful. false if failed.
2> *Rec: Found datacache record. Or nil if cache record isn't found.

Additional note:
- Method takes RD store-lock. Caller go-routine shouldn't invoke this method in any
store-lock. It's a deadlock otherwise.
- Returned record is locked. It's caller's responsibility to unlock the same.
**************************************************************************** */
func (pDataCache *DataCache) GetRec(key Key) (bool, *Rec) {
	if pDataCache == nil {
		return false, nil
	}

	pDataCache.ReadLock()
	defer pDataCache.ReadUnlock()

	pRec, isOK := pDataCache.cache[key]
	if !isOK {
		return false, nil
	}

	pRec.pRecLock.Lock()  // record is locked

	return true, pRec
}


/* ****************************************************************************
Description :
Function returns a locked datacache record. It's caller's responsibility to unlock the
same once done with it. Returned datacache record is a pointer to Rec.

Receiver    :
pDataCache *DataCache: Instance of datacache

Implements  : NA

Arguments   :
1> key Key: Key to the cache record.

Return value:
1> bool: true if successful. false if failed.
2> *Rec: Found datacache record. Or nil if cache record isn't found.

Additional note:
- Method doesn't take any store-lock and thus doesn't release any. Caller go-routine
must invoke this method in either WR or RD store-lock.
- Thus, the behaviour is unpredictable in case the calling go-routine calls this method
without any store-lock.
- Returned record is locked. It's caller's responsibility to unlock the same.
**************************************************************************** */
func (pDataCache *DataCache) GetRecWOLock(key Key) (bool, *Rec) {
	if pDataCache == nil {
		return false, nil
	}

	pRec, isOK := pDataCache.cache[key]
	if !isOK {
		return false, nil
	}

	pRec.pRecLock.Lock()  // record is locked

	return true, pRec
}


/* ****************************************************************************
Description :
Returns a copy of the data-payload of datacache record. It's a payload data and not the cache
record by itself. Therefore, locking and unlocking of cache record happens just through the method.

Receiver    :
pDataCache *DataCache: Instance of datacache.

Implements  : na

Arguments   :
1> key interface{}: Key to fetch cache record.

Return value:
1> bool: true if successful. returns false if datacache record for the given key
isn't found.
2> interface{}: Payload of fetched datacache record. It's a pointer. nil in case of error.

Additional note:
- Method takes RD store-lock. Caller go-routine shouldn't invoke this method in any
store-lock. It's a deadlock otherwise.
- Method returns payload of the cache record. Therefore, locking and unlocking of cache record
happens just through the method.
**************************************************************************** */
func (pDataCache *DataCache) GetDataRec(key Key) (bool, interface{}) {
	if pDataCache == nil {
		//return false, interface{}
		return false, nil
	}

	pDataCache.cacheLock.RLock()
	defer pDataCache.cacheLock.RUnlock()

	pRec, isOK := pDataCache.cache[key]
	if !isOK {
		//return false, interface{}
		return false, nil
	}

	pRec.pRecLock.Lock()
	pDataRec := pRec.PDataRec
	pRec.pRecLock.Unlock()

	return true, pDataRec
}


/* ****************************************************************************
Description :
Returns a copy of the data-payload of datacache record. It's a payload data and not the cache
record by itself. Therefore, locking and unlocking of cache record happens just through the method.

Receiver    :
pDataCache *DataCache: Instance of datacache.

Implements  : na

Arguments   :
1> key interface{}: Key to fetch cache record.

Return value:
1> bool: true if successful. false if failed. returns false if datacache record for the given key
isn't found.
2> interface{}: Payload of fetched datacache record. It's a pointer. nil in case of error.

Additional note:
- Method doesn't take any store-lock and thus doesn't release any. Caller go-routine must invoke this
method in either WR or RD store-lock.
- Thus, the behaviour is unpredictable in case the calling go-routine calls this method unguarded through
any store-lock.
- Method returns payload of the cache record and not the cache record by itself. Therefore, locking and
unlocking of cache record happens just through the method.
**************************************************************************** */
func (pDataCache *DataCache) GetDataRecWOLock(key Key) (bool, interface{}) {
	if pDataCache == nil {
		return false, nil
	}

	pRec, isOK := pDataCache.cache[key]
	if !isOK {
		return false, nil
	}

	pRec.pRecLock.Lock()
	pDataRec := pRec.PDataRec
	pRec.pRecLock.Unlock()

	return true, pDataRec
}


/* *****************************************************************************
Description :
Function checks if the given key exists in the cache or not. Returns true if it
exists.

Receiver    :
pDataCache *DataCache: Instance of datacache.

Arguments   :
1> key Key: Key to the cache record.

Return value:
1> bool: true if key exists. false if it doesn't.

Additional note: NA
***************************************************************************** */
func (pDataCache *DataCache) DoesKeyExist(key Key) bool {
	if pDataCache == nil {
		return false
	}

	pDataCache.ReadLock()
	defer pDataCache.ReadUnlock()

	if _, isOK := pDataCache.cache[key]; !isOK {
		return false
	}

	return true
}


/* *****************************************************************************
Description :
Method checks if the given key exists in the cache or not. Returns true if it
exists.

Receiver    :
pDataCache *DataCache: Instance of datacache.

Arguments   :
1> key Key: Key to the cache record.

Return value:
1> bool: true if key exists. false if it doesn't.

Additional note:
- As the name suggests, the method doesn't take any store lock. It rather assumes
it's invoked in the RD or WR store lock. Result is unpredictable in case it's not.
***************************************************************************** */
func (pDataCache *DataCache) DoesKeyExistWOLock(key Key) bool {
	if pDataCache == nil {
		return false
	}

	if _, isOK := pDataCache.cache[key]; !isOK {
		return false
	}

	return true
}


/* *****************************************************************************
Description :
Method returns number of records in the cache.

Receiver    :
pDataCache *DataCache: Datacache instance.

Implements  : NA

Arguments   : NA

Return value:
1> bool: true if successful. flase otherwise.
2> int: Number of records in the cache.

Additional note:
This method is guarded through WR store-lock and releases the same while returning.
Therefore, caller shouldn't take any lock before calling this method.
***************************************************************************** */
func (pDataCache *DataCache) GetCnt() (bool, int) {
	if pDataCache == nil {
		return false, 0
	}

	pDataCache.cacheLock.Lock()
	defer pDataCache.cacheLock.Unlock()

	return true, pDataCache.cnt
}


/* *****************************************************************************
Description :
Method sets cnt value and returns number of records in the cache.

Receiver    :
pDataCache *DataCache: Datacache instance.

Implements  : NA

Arguments   :
1> int val: Initial value of number of records in the cache.

Return value:
1> bool: true if successful. flase otherwise.
2> int: Number of records in the cache.

Additional note:
This method is guarded through WR store-lock and releases the same while returning.
Therefore, caller shouldn't take any lock before calling this method.
***************************************************************************** */
func (pDataCache *DataCache) SetCnt(val int) (bool, int) {
	if pDataCache == nil {
		return false, 0
	}

	pDataCache.cacheLock.Lock()
	defer pDataCache.cacheLock.Unlock()

	pDataCache.cnt = val

	return true, pDataCache.cnt
}


/* ****************************************************************************
Description :
Loads datache. Invokes loadfn() callback that actually performs the load operation.
Consumer of datacache passes this function while creating the datacache instance.

Receiver    :
pDataCache *DataCache: Instance of datacache.

Implements  : NA

Arguments   :
1> isLoaderProvided bool: Method reports error if this flag is true and load function
isn't provided.

Return value:
1> bool: true if successful, false otherwise.
2> error: Returns cause of error.

Additional note:
- Caller go-routine shouldn't invoke this method in any store-lock. It's deadlock in case it
does so.
- The method by itself takes WR store-lock and releases the same once the cache is loaded with
appropriate data.
**************************************************************************** */
func (pDataCache *DataCache) Load(isLoaderProvided bool) (bool, error) {
	var err error

	if pDataCache == nil {
		err = errors.New("Nil datacache.")
		return false, err
	}

	pDataCache.cacheLock.Lock()
	defer func() {
		if pDataCache.reciteratefn == nil {
			pDataCache.singletonFlag = true
		}
		pDataCache.cacheLock.Unlock()
	}()

	if pDataCache.singletonFlag {
		err = errors.New("Already executed load-time sequence.")
		return false, err
	}

	if pDataCache.loadfn == nil {
		if !isLoaderProvided {
			return true, nil
		}

		err = errors.New("Nil datacache loader.")
		return false, err
	}

	isOK, recList := pDataCache.loadfn()
	if !isOK {
		err = errors.New("Load function failed.")
		return false, err
	}

	for i, _ := range recList {
		pDataCacheRec := &Rec {
			PDataRec: recList[i].PDataRec,
			KeyList: recList[i].KeyList,
			isActive: true,
		}
		pDataCacheRec.pRecLock = &sync.Mutex{}
		//fmt.Printf("dbgrm::  dataRec: %#v\nkeyList: $#v\n", *pDataCacheRec.PDataRec, pDataCacheRec.KeyList)

		for j, _ := range recList[i].KeyList {
			pDataCache.cache[recList[i].KeyList[j]] = pDataCacheRec
		}
	}

	pDataCache.cnt = len(recList)

	return true, nil
}


/* ****************************************************************************
Description :
reciteratefn is the function provided by the consumer of the cache. Whilst the loop iterates
over each record in the cache, a function pointed by reciteratefn actually operates on each record.

Receiver    :
1> pDataCache *DataCache: DataCache store

Implements  : NA

Arguments   :
1> isIteratorProvided bool: Method reports error if this flag is true and iterator function
isn't provided.

Return value:
1> bool: true if successful, false otherwise.
2> error: Returns cause of error.

Additional note:
- Caller go-routine shouldn't invoke this method in any store-lock. It's deadlock in case it
does so.
- The method by itself takes WR store-lock and releases the same once done.
- Should be invoked only during server start-up and just after the cache is loaded.
and before any other subsystem, such as webserver, is initialized. That's the reason, each
iterated record isn't guarded in its own record lock.
**************************************************************************** */
func (pDataCache *DataCache) Iterate(cacheName string, isIteratorProvided bool) (bool, error) {
	var err error

	if pDataCache == nil {
		err = errors.New("Nil datacache.")
		return false, err
	}

	pDataCache.cacheLock.Lock()
	defer func() {
		pDataCache.singletonFlag = true
		pDataCache.cacheLock.Unlock()
	}()

	if pDataCache.singletonFlag {
		err = errors.New("Already executed load-time iteration sequence.")
		return false, err
	}

	if pDataCache.reciteratefn == nil {
		if !isIteratorProvided {
			return true, nil
		}

		err = errors.New("Nil datacache iterator.")
		return false, err
	}

	for _, pRec := range pDataCache.cache {
		//pRec.RecLock()
		pDataCache.reciteratefn(pRec.PDataRec)
		//pRec.RecUnlock()
	}

	return true, nil
}


/* *****************************************************************************
Description :
- Load and Iterate combined. Should be invoked only during server start-up and whilst
cache is being loaded for the first time.
- And particularly, before any other subsystem starts, for instance, webserver.
- singletonFlag is set to true in the end whilst still guarded by WR store-lock.

Receiver    :
pDataCache *DataCache: Instance of datacache.

Implements  : NA

Arguments   :
1> isLoaderProvided bool: Method reports error if this flag is true and load function
isn't provided.
2> isIteratorProvided bool: Method reports error if this flag is true and iterator function
isn't provided.

Return value:
1> bool: true if successful, false otherwise.
2> error: Returns cause of error.

Additional note:
- This method is a combination of Load() and Iterate() methods.
- Caller go-routine shouldn't invoke this method in any store-lock. It's deadlock in case it
does so.
- The method by itself takes WR store-lock and releases the same once done.
***************************************************************************** */
func (pDataCache *DataCache) LoadAndIterate(isLoaderProvided bool, isIteratorProvided bool) (bool, error) {
	var err error

	if pDataCache == nil {
		err = errors.New("Nil datacache.")
		return false, err
	}

	pDataCache.cacheLock.Lock()
	defer func() {
		pDataCache.singletonFlag = true
		pDataCache.cacheLock.Unlock()
	}()

	if pDataCache.singletonFlag {
		err = errors.New("Already executed all load-time sequences.")
		return false, err
	}

	if pDataCache.loadfn == nil {
		if !isLoaderProvided {
			return true, nil
		}

		err = errors.New("Nil datacache loader.")
		return false, err
	}

	isOK, recList := pDataCache.loadfn()
	if !isOK {
		err = errors.New("Load function failed.")
		return false, err
	}

	for i, _ := range recList {
		pDataCacheRec := &Rec {
			PDataRec: recList[i].PDataRec,
			KeyList: recList[i].KeyList,
			isActive: true,
		}
		pDataCacheRec.pRecLock = &sync.Mutex{}

		for j, _ := range recList[i].KeyList {
			pDataCache.cache[recList[i].KeyList[j]] = pDataCacheRec
		}
	}

	pDataCache.cnt = len(recList)

	if pDataCache.reciteratefn == nil {
		if !isIteratorProvided {
			return true, nil
		}

		err = errors.New("Nil datacache iterator.")
		return false, err
	}

	for _, pRec := range pDataCache.cache {
		pDataCache.reciteratefn(pRec.PDataRec)
	}

	return true, nil
}


/* ****************************************************************************
Description :
recHandler is the function provided by the consumer of the cache. Whilst the loop iterates
over each record in the cache, a function pointed by recHandler actually operates on each record.

Receiver    :
1> pDataCache *DataCache: DataCache store

Implements  : NA

Arguments   :
1> cacheName string: Just for a log message. Isn't being used right now.
2> recHandler RecHandlerFunc: Handler function of each iterated record.

Return value:
1> bool: true if successful, false otherwise.
2> error: Returns cause of error.

Additional note:
- Caller go-routine shouldn't invoke this method in any store-lock. It's deadlock in case it
does so.
- The method by itself takes WR store-lock and releases the same once done. Also, recHandler
is well guarded in the record lock of each iterated record.
- Addional word of caution:
This method shouod be invoked knowing that it's going to cause a performance issue in the running
server as it holds WR store-lock and each iterated record is guarded in its own record lock.
**************************************************************************** */
func (pDataCache *DataCache) AuxIterate(cacheName string, recHandler RecHandlerFunc) (bool, error) {
	var err error

	if pDataCache == nil {
		err = errors.New("Nil datacache.")
		return false, err
	}

	if recHandler == nil {
		err = errors.New("Nil or no cache record handler provided.")
		return false, err
	}

	pDataCache.WriteLock()
	defer pDataCache.WriteUnlock()

	if recHandler == nil {
		err = errors.New("Nil datacache iterator.")
		return false, err
	}

	for _, pRec := range pDataCache.cache {
		pRec.RecLock()
		recHandler(pRec)
		pRec.RecUnlock()
	}

	return true, nil
}


/* ****************************************************************************
Description :
Function creates datacache instance.

Receiver    : NA

Implements  : NA

Arguments   :
1> loadFunc LoadFunc: Callback used by Load() to load the cache.
Each cache may have its own specific way of loading the cache. Load() invokes
the function referred by loadFunc to load the cache.
2> iteratorFunc RecHandlerFunc: Callback used by Iterate().
iteratorFunc is invoked on payload of each iterated datacache record.

Return value:
1> *DataCache: Newly created datacache instance.

Additional note: NA
**************************************************************************** */
func Create(loadFunc LoadFunc, iteratorFunc RecHandlerFunc) *DataCache {
	pDataCache := &DataCache {
		cache: make(CacheStore),
		loadfn: loadFunc,
		reciteratefn: iteratorFunc,
	}

	return pDataCache
}

package cache_golang

import (
	"sync"
	"time"
	"github.com/pkg/errors"
)

/*
 Structure that must be embedded in the object that should be cached with expiration
 If no expiration is desired this can be ignored
 */
type XEntry struct {
	sync.Mutex                   //互斥锁
	key            string        //key
	keepAlive      bool          //是否保持连接
	expireDuration time.Duration //到期时间
	expiringSince  time.Time     //时间
}

/**
 The private interface
 */
type expiringCacheEntry interface {
	XCache(key string, expire time.Duration, value expiringCacheEntry)
	KeepAlive()
	ExpiringSince() time.Time
	ExpireDuration() time.Duration
}

var (
	xcache              = make(map[string]expiringCacheEntry)
	cache               = make(map[string]interface{})
	xMux                sync.RWMutex
	mux                 sync.RWMutex
	expTimer            *time.Timer
	ExpireCheckInterval = 30 * time.Second
)

/**
  init
 */
func Init() {
	expirationCheck()
}

/**
   check
 */
func expirationCheck() {
	go removeExpiredEntries()
}

/**
  remove
 */
func removeExpiredEntries() {
	//Take a copy of xcache so we can iterate over it without blocking the mutex
	xMux.Lock()
	cache := xcache
	xMux.Unlock()

	now := time.Now()
	for key, c := range cache {
		if now.Sub(c.ExpiringSince()) >= c.ExpireDuration() {
			xMux.Lock()
			delete(xcache, key)
			xMux.Unlock()
		}
	}
	expTimer = time.AfterFunc(ExpireCheckInterval, expirationCheck)
}

/**
  The main function to cache with expiration
 */
func (xe *XEntry) XCache(key string, expire time.Duration, value expiringCacheEntry) {
	xe.keepAlive = true
	xe.key = key
	xe.expireDuration = expire
	xe.expiringSince = time.Now()

	xMux.Lock()
	defer xMux.Unlock()
	xcache[key] = value
}

/**
   Mark entry to be kept for another expirationDuration period
 */
func (xe *XEntry) KeepAlive() {
	xe.Lock()
	defer xe.Unlock()
	xe.expiringSince = time.Now()
}

/**
  Returns this entry's expiration duration
 */
func (xe *XEntry) ExpireDuration() time.Duration {
	xe.Lock()
	defer xe.Unlock()
	return xe.expireDuration
}

/**
  Returns since when this entry is expiring
 */
func (xe *XEntry) ExpiringSince() time.Time {
	xe.Lock()
	defer xe.Unlock()
	return xe.expiringSince
}

/**
   The function to be used to cache a key/value pair when expiration is not needed
 */
func Cache(key string, value interface{}) {
	mux.Lock()
	defer mux.Unlock()
	cache[key] = value
}

/**
  Get an entry from the expiration cache and mark it to be kept alive
 */
func GetXCached(key string) (ece expiringCacheEntry, err error) {
	xMux.RLock()
	defer xMux.RUnlock()
	if r, ok := xcache[key]; ok {
		r.KeepAlive() //Start calling connection
		return r, nil
	}
	return nil, errors.New("not found")
}

/**
   The function to extract a value for a key that never expires
 */
func GetCached(key string) (v interface{}, err error) {
	mux.RLock()
	defer mux.RUnlock()
	if r, ok := cache[key]; ok {
		return r, nil
	}
	return nil, errors.New("not found")
}

/**
  Delete all keys from expiration cache
 */
func Flush() {
	mux.Lock()
	defer mux.Unlock()
	cache = make(map[string]interface{})
}

/**
  Delete all keys from cache
 */
func XFulsh() {
	xMux.Lock()
	defer xMux.Unlock()
	xcache = make(map[string]expiringCacheEntry)
}

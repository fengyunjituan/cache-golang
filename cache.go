package cache_golang

import (
	"sync"
	"time"
	"github.com/pkg/errors"
)

/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

/*
 Structure that must be embeded in the objectst that must be cached with expiration.
 If the expiration is not needed this can be ignored
 */
type XEntry struct {
	sync.Mutex                   //互斥锁
	key            string        //key
	keepAlive      bool          //是否保持连接
	expireDuration time.Duration //到期时间
}

/**
 The private interface
 */
type expiringCacheEntry interface {
	XCache(key string, expire time.Duration, value expiringCacheEntry)
	Expire()
	KeepAlive()
}

var (
	xcache = make(map[string]expiringCacheEntry)
	cache  = make(map[string]interface{})
)

/**
  The main function to cache with expiration
 */
func (xe *XEntry) XCache(key string, expire time.Duration, value expiringCacheEntry) {
	xe.keepAlive = true
	xe.key = key
	xe.expireDuration = expire
	xcache[key] = value
	go xe.Expire()
}

/**
  The internal mechanism for expiartion
 */
func (xe *XEntry) Expire() {
	for xe.keepAlive {
		xe.Lock()
		xe.keepAlive = false
		xe.Unlock()
		t := time.NewTimer(xe.expireDuration) //设置到期时间
		<-t.C
		xe.Lock()
		if !xe.keepAlive {
			delete(xcache, xe.key)
		}
		xe.Unlock()
	}
}

/**
   Mark entry to be kept another expirationDuration period
 */
func (xe *XEntry) KeepAlive() {
	xe.Lock()
	defer xe.Unlock()
	xe.keepAlive = true
}

/**
   The function to be used to cache a key/value pair when expiration is not needed
 */
func Cache(key string, value interface{}) {
	cache[key] = value
}

/**
  Get an entry from the expiration cache and mark it for keeping alive
 */
func GetXCached(key string) (ece expiringCacheEntry, err error) {
	//Determine whether the key is in the map
	if r, ok := xcache[key]; ok {
		r.KeepAlive() //Start calling connection
		return r, nil
	}
	return nil, errors.New("not found")
}

/**
   The function to extract a value for a key that never expire
 */
func GetCached(key string) (v interface{}, err error) {
	if r, ok := cache[key]; ok {
		return r, nil
	}
	return nil, errors.New("not found")
}

/**
   Delete all keys from expiration cache
 */
func (xe *XEntry) Flush() {
	xcache = make(map[string]expiringCacheEntry)
}

/**
  Delete all keys from cache
 */
func Flush() {
	cache = make(map[string]interface{})
}

package cache_golang

import (
	"testing"
	"time"
)

type myCache struct {
	XEntry
	data string
}

func init() {
	ExpireCheckInterval = 1 * time.Second
}
func TestCache(t *testing.T) {
	a := &myCache{data: "feng qi yun yong"}
	a.XCache("feng", 1*time.Second, a, nil)
	b, err := GetXCached("feng")
	if err != nil || b == nil || b != a {
		t.Error("Error retriving data from cache", err)
	}
}

func TestCacheExpire(t *testing.T) {
	a := &myCache{data: "feng qi yun yong"}
	a.XCache("feng", 1*time.Second, a, nil)
	b, err := GetXCached("feng")
	if err != nil || b == nil || b.(*myCache).data != "feng qi yun yong" {
		t.Error("Error retriving data from cache", err)
	}
	time.Sleep(2001 * time.Microsecond)
	b, err = GetXCached("feng")
	if err == nil || b != nil {
		t.Error("Error expiring data")
	}
}

func TestCacheKeepAlive(t *testing.T) {
	a := &myCache{data: "feng qi yun yong"}
	a.XCache("feng", 1*time.Second, a, nil)
	b, err := GetXCached("feng")
	if err != nil || b == nil || b.(*myCache).data != "feng qi yun yong" {
		t.Error("Error retriving data from cache", err)
	}
	time.Sleep(500 * time.Millisecond)
	b.KeepAlive()
	time.Sleep(2001 * time.Millisecond)
	if err != nil {
		t.Error("Error keeping cache data alive", err)
	}
	time.Sleep(1000 * time.Millisecond)
	b, err = GetXCached("feng")
	if err == nil || b != nil {
		t.Error("Error expiring data")
	}
}

func TestFlush(t *testing.T) {
	a := &myCache{data: "feng qi yun yong"}
	a.XCache("feng", 10*time.Second, a, nil)
	time.Sleep(1000 * time.Millisecond)
	XFulsh()
	b, err := GetXCached("feng")
	if err == nil || b != nil {
		t.Error("Error expiring data")
	}
}

func TestFlushNoTimeout(t *testing.T) {
	a := &myCache{data: "feng qi yun yong"}
	a.XCache("feng", 10*time.Second, a, nil)
	XFulsh()
	b, err := GetXCached("feng")
	if err == nil || b != nil {
		t.Error("Error expiring data")
	}
}

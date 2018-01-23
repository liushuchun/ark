package urlfetch

import (
	"strings"
	"testing"
	"time"
)

func TestExpireTime(t *testing.T) {
	cache := NewLocalCache("/tmp", 10000*time.Millisecond)
	f := strings.NewReader("abcdefg")
	key := "test"
	defer cache.Delete(key)

	_, err := cache.Set(key, f)
	if err != nil {
		t.Fatalf("set %s failed", key, err)
	}

	_, err = cache.Get(key)
	if err != nil {
		t.Fatalf("get %s failed, %v", key, err)
	}

	cache.ExpireTime = 99 * time.Millisecond
	time.Sleep(100 * time.Millisecond)
	_, err = cache.Get(key)
	if err != ErrExpired {
		t.Fatalf("expect %v, but got %v", ErrExpired, err)
	}
}

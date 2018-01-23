package trace

import (
	"fmt"
	"testing"
	"time"
)

func TestTokenRateSampler(t *testing.T) {
	sam := NewTokenRateSampler(100)

	go func() {
		for {
			time.Sleep(1e5)
			sam.Sample()
		}
	}()
	go func() {
		for {
			time.Sleep(1e5)
			sam.Sample()
		}
	}()

	t1 := time.Now().UnixNano()
	for i := 0; i < 1000000; i++ {
		sam.Sample()
	}
	t2 := time.Now().UnixNano()
	fmt.Println(t2-t1, "ns")

	t1 = time.Now().UnixNano()
	for i := 0; i < 1000000; i++ {
		sam.Sample()
	}
	t2 = time.Now().UnixNano()
	fmt.Println(t2-t1, "ns")
}

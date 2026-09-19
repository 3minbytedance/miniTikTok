package common

import (
	"sync"
	"testing"
)

func TestInitSnowflakeAndGetUid(t *testing.T) {
	if err := InitSnowflake(1); err != nil {
		t.Fatalf("InitSnowflake 失败: %v", err)
	}

	// 并发生成，验证唯一性
	const goroutines = 100
	const per = 100
	seen := make(chan uint, goroutines*per)
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < per; j++ {
				seen <- GetUid()
			}
		}()
	}
	wg.Wait()
	close(seen)

	uniq := make(map[uint]struct{}, goroutines*per)
	for id := range seen {
		if _, dup := uniq[id]; dup {
			t.Fatalf("生成重复 ID: %d", id)
		}
		uniq[id] = struct{}{}
	}
	if len(uniq) != goroutines*per {
		t.Errorf("生成数量 = %d, want %d", len(uniq), goroutines*per)
	}
}

func TestGetUidMonotonic(t *testing.T) {
	if err := InitSnowflake(1); err != nil {
		t.Fatalf("InitSnowflake 失败: %v", err)
	}
	prev := GetUid()
	for i := 0; i < 1000; i++ {
		next := GetUid()
		if next <= prev {
			t.Fatalf("ID 应单调递增: prev=%d next=%d", prev, next)
		}
		prev = next
	}
}

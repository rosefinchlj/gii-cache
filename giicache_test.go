package giicache

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	gii := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB](callback) search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				} else {
					loadCounts[key] += 1
				}

				return []byte(v), nil
			}

			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range db {
		// 第一次访问，因为没有缓存，所以会触发回调函数
		// 就是常说的缓存击穿
		if view, err := gii.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value: ", err)
		}

		// 第二次访问，缓存可以命中，不会触发回调函数, loadCounts[k]为0
		if _, err := gii.Get(k); err != nil || loadCounts[k] > 0 {
			t.Fatalf("cache %s miss", k)
		}
	}

	// 说的是缓存穿透
	if view, err := gii.Get("unknown"); err == nil {
		t.Fatalf("he value of unknow should be empty, but %s got ", view)
	}
}

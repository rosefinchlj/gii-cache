package giicache

import (
	"github.com/gii-cache/lru"
	"log"
	"sync"
)

type cache struct {
	sync.Mutex

	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.Lock()
	defer c.Unlock()

	// Lazy initialize，第一次使用的时候才会分配内存初始化
	if c.lru == nil {
		log.Println("during add cache, lru init")
		c.lru = lru.New(c.cacheBytes, nil)
	}

	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.Lock()
	defer c.Unlock()

	if c.lru == nil {
		log.Printf("cache is nil")
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}

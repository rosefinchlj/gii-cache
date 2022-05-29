package giicache

import (
	"errors"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (g GetterFunc) Get(key string) ([]byte, error) {
	return g(key)
}

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	mainCache cache
	getter    Getter
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		mainCache: cache{cacheBytes: cacheBytes},
		getter:    getter,
	}

	groups[name] = g

	return g
}

func GetGroup(name string) *Group {
	mu.RLock() // 读写锁，这里不涉及到写操作，只是读操作
	defer mu.RUnlock()

	return groups[name]
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Printf("[%s] hit", key)
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) // 回调用户自定义的函数，逻辑处理由用户自定义
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)

	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

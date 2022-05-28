package lru

import "container/list"

// LRU(Least Recently Used)
// 最近最少使用，相对于仅考虑时间因素的 FIFO 和仅考虑访问频率的 LFU，LRU 算法可以认为是相对平衡的一种淘汰算法。
// LRU 认为，如果数据最近被访问过，那么将来被访问的概率也会更高。LRU 算法的实现非常简单，维护一个队列，
// 如果某条记录被访问了，则移动到队首，
// 那么队尾则是最近最少访问的数据，淘汰该条记录即可

type Value interface {
	Len() int // Return the length of the value
}

type entry struct {
	key   string
	value Value
}

type Cache struct {
	maxBytes  int64 // 最大缓存字节数
	nBytes    int64 // 当前缓存字节数
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) // 当需要清除一个缓存时回调
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) // 移到队首
		kv := ele.Value.(*entry)
		return kv.value, true
	}

	return
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 当前缓存字节数减去被删除的 key 和 value 的长度
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value) // 回调用户注册的函数
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) // 这里也是访问了key，移到队首
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len()) // 当前缓存字节数减去旧的 value 的长度，加上新的 value 的长度
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len()) // 当前缓存字节数增加 key 和 value 的长度
	}

	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

// Len 为了方便测试，我们实现 Len() 用来获取添加了多少条数据
func (c *Cache) Len() int {
	return c.ll.Len()
}

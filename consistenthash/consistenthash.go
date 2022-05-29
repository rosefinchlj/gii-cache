package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// Map 包含所有节点的hash值
type Map struct {
	hash     Hash
	replicas int            // 虚拟节点数
	keys     []int          // 哈希环上所有的key  Sorted
	hashMap  map[int]string // 虚拟节点与真是节点的映射
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(key + ":" + strconv.Itoa(i)))) // 节点A-1 节点A-2
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	// 对keys进行排序，便于Get方法进行二分查找
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	// 二分查找
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}

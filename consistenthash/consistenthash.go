package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// Map 包含所有节点的hash值
type Map struct {
	hashFun  Hash
	replicas int            // 虚拟节点数
	keys     []int          // 哈希环上所有的key  Sorted
	hashMap  map[int]string // 虚拟节点与真是节点的映射
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hashFun:  fn,
		hashMap:  make(map[int]string),
	}

	if m.hashFun == nil {
		m.hashFun = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加节点
func (m *Map) Add(nodes ...string) {
	for _, node := range nodes {
		for i := 0; i < m.replicas; i++ {
			hashedNode := int(m.hashFun([]byte(strconv.Itoa(i) + node)))
			m.keys = append(m.keys, hashedNode)
			m.hashMap[hashedNode] = node
		}
	}
	// 对keys进行排序，便于Get方法进行二分查找
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hashFun([]byte(key)))
	// 二分查找，没有找到返回n [0, n)，不是返回-1
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// 如果 idx == len(m.keys)，说明应选择 m.keys[0]，因为 m.keys 是一个环状结构，所以用取余数的方式来处理这种情况。
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

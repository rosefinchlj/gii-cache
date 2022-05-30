package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// add后虚拟节点为：2, 4, 6, 12, 14, 16, 22, 24, 26
	// [2 12 22]对应真实节点 2
	// [4 14 24]对应真实节点 4
	// [6 16 26]对应真实节点 6
	hash.Add("6", "4", "2")

	// key的hash算法为：把字符串转换为整数
	// 所以：k通过hash后，对虚拟节点数取余
	testCases := map[string]string{
		"2":  "2", // 虚拟节点数索引0的虚拟节点2 对应真实节点 2
		"11": "2", // 虚拟节点数索引3(11 <= 12,顺时针取第一个满足条件的节点)的虚拟节点12 对应真实节点 2
		"13": "4", // 虚拟节点数索引4(13 <= 14,顺时针取第一个满足条件的节点)的虚拟节点14 对应真实节点 4
		"19": "2", // 虚拟节点数索引5(19 <= 22,顺时针取第一个满足条件的节点)的虚拟节点22 对应真实节点 2
		"27": "2", // 虚拟节点数索引5(20 <= 22,顺时针取第一个满足条件的节点)的虚拟节点22 对应真实节点 2
	}

	for k, v := range testCases {
		if hash := hash.Get(k); hash != v {
			t.Errorf("Asking for %s, should have yielded %s, but yielded %s", k, hash, v)
		}
	}

	// 增加节点
	// 8 18 28
	hash.Add("8")

	// 落在节点 8上
	testCases["27"] = "8"

	for k, v := range testCases {
		if hash := hash.Get(k); hash != v {
			t.Errorf("Asking for %s, should have yielded %s, but yielded %s", k, hash, v)
		}
	}
}

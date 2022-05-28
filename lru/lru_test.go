package lru

import (
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestCache_Get(t *testing.T) {
	lru := New(0, nil)
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatal("cache hit key1 = 123")
	}
}

func TestCache_RemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"

	c := len(k1 + k2 + v1 + v2) // 只初始化两个元素的大小
	lru := New(int64(c), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3)) // 这里会触发移除k1(满足最近最少使用) k1 -> k2 -> k3

	if lru.Len() != 2 { // k1被移除
		t.Fatalf("bad len: %d", lru.Len())
	}

	if _, ok := lru.Get(k1); ok {
		t.Fatalf("should not contain k1")
	}

	if _, ok := lru.Get(k2); !ok {
		t.Fatalf("should contain k2")
	}

	if _, ok := lru.Get(k3); !ok {
		t.Fatalf("should contain k3")
	}
}

func TestCache_OnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		//t.Log("remove key:", key)
		keys = append(keys, key)
	}

	lru := New(int64(10), callback)
	lru.Add("key1", String("12356")) // 这里达到缓存最大值
	lru.Add("key2", String("456"))   // add key2, key1被移除
	lru.Add("key3", String("789"))   // add key3, key2被移除
	lru.Add("key4", String("111"))   // add key4, key3被移除

	expect := []string{"key1", "key2", "key3"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("expect keys %v, got %v", expect, keys)
	}

	value, ok := lru.Get("key4")
	if !ok || string(value.(String)) != "111" {
		t.Fatalf("expect key4's value 111, got %v", value)
	}
}

package singleflight

import "sync"

type call struct {
	sync.WaitGroup

	val interface{}
	err error
}

type Group struct {
	sync.Mutex
	m map[string]*call
}

func (g Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.Unlock()
		c.Wait() // 如果请求正在进行中，则等待

		return c.val, c.err // 请求结束，返回结果
	}

	// 不存在
	c := new(call)
	c.Add(1)     // 发起请求前加锁
	g.m[key] = c // 添加到 g.m，表明 key 已经有对应的请求在处理
	g.Unlock()

	c.val, c.err = fn()
	c.Done()

	g.Lock()
	delete(g.m, key)
	g.Unlock()

	return c.val, c.err
}

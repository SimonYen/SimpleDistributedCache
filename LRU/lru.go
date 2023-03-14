package lru

import "container/list"

type Cache struct {
	maxBytes   uint64 //允许使用的最大内存
	nBytes     uint64 //当前已使用的内存
	doubleList *list.List
	cache      map[string]*list.Element
	OnEvicted  func(key string, value Value) //回调函数
}

func New(maxBytes uint64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:   maxBytes,
		doubleList: list.New(),
		cache:      make(map[string]*list.Element),
		OnEvicted:  onEvicted,
	}
}

type Value interface {
	Len() int //返回值所占的内存大小
}

type entry struct {
	key   string
	value Value
}

// 查找功能
func (c *Cache) Get(key string) (value Value, ok bool) {
	if e, ok := c.cache[key]; ok {
		c.doubleList.MoveToFront(e)
		kv := e.Value.(*entry)
		return kv.value, true
	}
	return
}

// 删除功能
func (c *Cache) RemoveOldest() {
	e := c.doubleList.Back()
	//链表不为空的话
	if e != nil {
		c.doubleList.Remove(e)
		kv := e.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= uint64(len(kv.key)) + uint64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 新增或覆盖功能
func (c *Cache) Add(key string, value Value) {
	if e, ok := c.cache[key]; ok {
		//覆盖情况
		c.doubleList.MoveToFront(e)
		kv := e.Value.(*entry)
		c.nBytes += uint64(value.Len()) - uint64(kv.value.Len())
		kv.value = value
	} else {
		//新增情况
		e := c.doubleList.PushFront(&entry{key: key, value: value})
		c.cache[key] = e
		c.nBytes += uint64(value.Len()) + uint64(len(key))
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

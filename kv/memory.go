package kv

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type Memory struct {
	cache *cache.Cache
}

func NewMemory() *Memory {
	return &Memory{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (m *Memory) Set(key string, o interface{}, ttl time.Duration) error {
	m.cache.Set(key, o, ttl)
	return nil
}

func (m *Memory) Get(key string) (interface{}, error) {
	if o, found := m.cache.Get(key); found {
		return o, nil
	}
	return nil, ErrKeyMiss
}

func (m *Memory) Del(key string) error {
	m.cache.Delete(key)
	return nil
}

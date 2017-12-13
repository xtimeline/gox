package kv

import (
	"time"

	cache "gopkg.in/go-redis/cache.v5"
	redis "gopkg.in/redis.v5"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

type Redis struct {
	codec *cache.Codec
}

func NewRedisCluster(addrs []string) *Redis {
	codec := &cache.Codec{
		Redis: redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        addrs,
			PoolSize:     512,
			PoolTimeout:  10 * time.Second,
			IdleTimeout:  10 * time.Second,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}),
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	return &Redis{codec: codec}
}

func NewRedis(addr string) *Redis {
	codec := &cache.Codec{
		Redis: redis.NewClient(&redis.Options{
			DB:           0,
			Addr:         addr,
			PoolSize:     512,
			PoolTimeout:  10 * time.Second,
			IdleTimeout:  10 * time.Second,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}),
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	return &Redis{codec: codec}
}

func (r *Redis) Set(key string, o interface{}, ttl time.Duration) error {
	return r.codec.Set(&cache.Item{
		Key:        key,
		Object:     o,
		Expiration: ttl,
	})
}

func (r *Redis) Get(key string, o interface{}) error {
	if err := r.codec.Get(key, o); err != nil {
		if err == cache.ErrCacheMiss {
			return ErrKeyMiss
		}
		return err
	}

	return nil
}

func (r *Redis) Del(key string) error {
	if err := r.codec.Delete(key); err != nil {
		if err == cache.ErrCacheMiss {
			return nil
		}
		return err
	}
	return nil
}

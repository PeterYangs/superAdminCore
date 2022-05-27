package redisCache

import (
	"context"
	"github.com/PeterYangs/superAdminCore/v2/conf"
	"github.com/PeterYangs/superAdminCore/v2/redis"
	"time"
)

type redisCache struct {
}

func NewRedisCache() *redisCache {

	return &redisCache{}
}

func (r *redisCache) Put(key string, value string, ttl time.Duration) error {
	//TODO implement me
	_, err := redis.GetClient().Set(context.Background(), conf.Get("cache_prefix").(string)+":"+key, value, ttl).Result()

	return err
}

func (r *redisCache) Get(key string) (string, error) {
	//TODO implement me
	//panic("implement me")

	return redis.GetClient().Get(context.Background(), conf.Get("cache_prefix").(string)+":"+key).Result()

}

func (r *redisCache) Remove(key string) error {

	_, err := redis.GetClient().Del(context.Background(), conf.Get("cache_prefix").(string)+":"+key).Result()

	return err
}

func (r *redisCache) Exists(key string) bool {

	num, err := redis.GetClient().Exists(context.Background(), conf.Get("cache_prefix").(string)+":"+key).Result()

	if err != nil {

		return false
	}

	if num == 0 {

		return false
	}

	return true

}

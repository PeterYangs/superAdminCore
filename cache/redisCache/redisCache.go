package redisCache

import (
	"context"
	"github.com/PeterYangs/superAdminCore/conf"
	"github.com/PeterYangs/superAdminCore/redis"
	"time"
)

type redisCache struct {
}

func NewRedisCache() *redisCache {

	return &redisCache{}
}

func (r redisCache) Put(key string, value string, ttl time.Duration) error {
	//TODO implement me
	_, err := redis.GetClient().Set(context.TODO(), conf.Get("cache_prefix").(string)+":"+key, value, ttl).Result()

	return err
}

func (r redisCache) Get(key string) (string, error) {
	//TODO implement me
	//panic("implement me")

	return redis.GetClient().Get(context.TODO(), conf.Get("cache_prefix").(string)+":"+key).Result()

}

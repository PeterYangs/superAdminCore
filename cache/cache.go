package cache

import (
	"github.com/PeterYangs/superAdminCore/cache/fileCache"
	"github.com/PeterYangs/superAdminCore/cache/redisCache"
	"os"
	"time"
)

type CacheContract interface {
	// Put 过去时间为0为不过期
	Put(key string, value string, ttl time.Duration) error
	Get(key string) (string, error)
	Exists(key string) bool
	Remove(key string) error
}

func Cache() CacheContract {

	driver := os.Getenv("CACHE_DRIVER")

	switch driver {

	case "redis":

		return redisCache.NewRedisCache()

	case "file":

		return fileCache.NewFileCache()

	}

	return redisCache.NewRedisCache()
}

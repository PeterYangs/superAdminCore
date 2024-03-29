package redis

import (
	"context"
	"github.com/PeterYangs/superAdminCore/v2/component/logs"
	"github.com/PeterYangs/superAdminCore/v2/conf"
	"github.com/go-redis/redis/v8"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cast"
	"os"
	"sync"
	"time"
)

//var client *redis.Client

var once = sync.Once{}

type _connect struct {
	connect *redis.Client
}

type lock struct {
	key        string
	expiration time.Duration
	connect    *_connect
	requestId  string
	//checkCancel chan bool
	mu         sync.Mutex
	cancelCxt  context.Context
	cancelFunc context.CancelFunc
}

var Client *_connect

func connect() {

	cxt, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	conf_ := &redis.Options{
		Addr:         os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		DB:           0,
		MinIdleConns: cast.ToInt(os.Getenv("REDIS_MIN_IDLE_CONNS")),
		PoolSize:     cast.ToInt(os.Getenv("REDIS_POOL_SIZE")),
	}

	if os.Getenv("REDIS_PASSWORD") != "null" {

		conf_.Password = os.Getenv("REDIS_PASSWORD")
	}

	c := redis.NewClient(conf_)

	re := c.Ping(cxt)

	defer cancel()

	if re.Err() != nil {

		//panic(re.Err())

		logs.NewLogs().Error(re.Err().Error())

		//return

	}

	Client = &_connect{
		connect: c,
	}

}

func GetClient() *_connect {

	if Client == nil {

		once.Do(func() {

			connect()
		})

	}

	return Client

}

func (cc *_connect) Ping(cxt context.Context) *redis.StatusCmd {

	return cc.connect.Ping(cxt)
}

func (cc *_connect) Get(cxt context.Context, key string) *redis.StringCmd {

	return cc.connect.Get(cxt, conf.Get("redis_prefix").(string)+key)
}

// Set expiration为0则为不过期
func (cc *_connect) Set(cxt context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {

	return cc.connect.Set(cxt, conf.Get("redis_prefix").(string)+key, value, expiration)
}

// Exists key是否存在，返回0则一个都不存在，返回2这则有两个
func (cc *_connect) Exists(cxt context.Context, keys ...string) *redis.IntCmd {

	temp := make([]string, len(keys))

	for i, key := range keys {

		temp[i] = conf.Get("redis_prefix").(string) + key
	}

	return cc.connect.Exists(cxt, temp...)
}

func (cc *_connect) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {

	return cc.connect.Expire(ctx, conf.Get("redis_prefix").(string)+key, expiration)
}

func (cc *_connect) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {

	return cc.connect.SetNX(ctx, conf.Get("redis_prefix").(string)+key, value, expiration)

}

func (cc *_connect) Del(ctx context.Context, keys ...string) *redis.IntCmd {

	temp := make([]string, len(keys))

	for i, key := range keys {

		temp[i] = conf.Get("redis_prefix").(string) + key
	}

	return cc.connect.Del(ctx, temp...)
}

// LPush 列表，头部插入
func (cc *_connect) LPush(cxt context.Context, key string, value ...interface{}) *redis.IntCmd {

	return cc.connect.LPush(cxt, conf.Get("redis_prefix").(string)+key, value)

}

// RPush 列表，尾部插入
func (cc *_connect) RPush(cxt context.Context, key string, value ...interface{}) *redis.IntCmd {

	return cc.connect.RPush(cxt, conf.Get("redis_prefix").(string)+key, value)

}

func (cc *_connect) RPop(cxt context.Context, key string) *redis.StringCmd {

	return cc.connect.RPop(cxt, conf.Get("redis_prefix").(string)+key)
}

// BRPop 移除列表最后一个元素（阻塞，可监听多个队列）
func (cc *_connect) BRPop(cxt context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd {

	temp := make([]string, len(keys))

	for i, key := range keys {

		temp[i] = conf.Get("redis_prefix").(string) + key
	}

	return cc.connect.BRPop(cxt, timeout, temp...)
}

// BLPop 移除列表第一个元素（阻塞，可监听多个队列）
func (cc *_connect) BLPop(cxt context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd {

	temp := make([]string, len(keys))

	for i, key := range keys {

		temp[i] = conf.Get("redis_prefix").(string) + key
	}

	return cc.connect.BLPop(cxt, timeout, temp...)
}

func (cc *_connect) LRange(cxt context.Context, key string, start, stop int64) *redis.StringSliceCmd {

	return cc.connect.LRange(cxt, conf.Get("redis_prefix").(string)+key, start, stop)
}

// LLen 获取列表长度
func (cc *_connect) LLen(cxt context.Context, key string) *redis.IntCmd {

	return cc.connect.LLen(cxt, conf.Get("redis_prefix").(string)+key)
}

func (cc *_connect) ZAdd(cxt context.Context, key string, members ...*redis.Z) *redis.IntCmd {

	//fmt.Println(conf.Get("redis_prefix"))

	return cc.connect.ZAdd(cxt, conf.Get("redis_prefix").(string)+key, members...)
}

func (cc *_connect) ZRangeByScore(cxt context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {

	return cc.connect.ZRangeByScore(cxt, conf.Get("redis_prefix").(string)+key, opt)
}

func (cc *_connect) ZRange(cxt context.Context, key string, start, stop int64) *redis.StringSliceCmd {

	return cc.connect.ZRange(cxt, conf.Get("redis_prefix").(string)+key, start, stop)
}

func (cc *_connect) ZCard(cxt context.Context, key string) *redis.IntCmd {

	return cc.connect.ZCard(cxt, conf.Get("redis_prefix").(string)+key)
}

func (cc *_connect) ZRemRangeByRank(cxt context.Context, key string, start, stop int64) *redis.IntCmd {

	return cc.connect.ZRemRangeByRank(cxt, conf.Get("redis_prefix").(string)+key, start, stop)
}

func (cc *_connect) HMSet(cxt context.Context, key string, values ...interface{}) *redis.BoolCmd {

	return cc.connect.HMSet(cxt, conf.Get("redis_prefix").(string)+key, values...)
}

func (cc *_connect) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {

	return cc.connect.Eval(ctx, script, keys, args...)
}

// Lock 声明锁
func (cc *_connect) Lock(key string, expiration time.Duration) *lock {

	lockId := uuid.NewV4().String()

	//取消提醒context
	cxt, cancelCxt := context.WithCancel(context.Background())

	return &lock{
		key:        conf.Get("lock_prefix").(string) + key,
		expiration: expiration,
		connect:    cc,
		requestId:  lockId,
		mu:         sync.Mutex{},
		cancelCxt:  cxt,
		cancelFunc: cancelCxt,
	}
}

// Get 获取锁
func (lk *lock) Get() bool {

	cxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	ok, err := lk.connect.SetNX(cxt, lk.key, lk.requestId, lk.expiration).Result()

	if err != nil {

		return false
	}

	if ok {

		//检查锁是否过期
		go lk.checkLockIsRelease()

	}

	return ok
}

// Block 获取锁阻塞
func (lk *lock) Block(expiration time.Duration) bool {

	t := time.Now()

	for {

		cxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

		ok, err := lk.connect.SetNX(cxt, lk.key, lk.requestId, lk.expiration).Result()

		cancel()

		if err != nil {

			return false
		}

		if ok {

			go lk.checkLockIsRelease()

			return true
		}

		time.Sleep(200 * time.Millisecond)

		if time.Now().Sub(t) > expiration {

			return false
		}

	}

}

// Release 释放锁
func (lk *lock) Release() error {

	lk.mu.Lock()

	defer lk.mu.Unlock()

	cxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	const luaScript = `
	if redis.call('get', KEYS[1])==ARGV[1] then
		return redis.call('del', KEYS[1])
	else
		return 0
	end
	`
	res, err := redis.NewScript(luaScript).Run(cxt, GetClient().connect, []string{conf.Get("redis_prefix").(string) + lk.key}, lk.requestId).Result()

	//lk.mu.Unlock()

	if err != nil {

		return err
	}

	if res.(int64) != 0 {

		lk.cancelFunc()

	}

	return err

}

// ForceRelease 强制释放锁，忽略请求id
func (lk *lock) ForceRelease() error {

	lk.mu.Lock()

	defer func() {

		lk.cancelFunc()

		lk.mu.Unlock()

	}()

	cxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	_, err := lk.connect.Del(cxt, lk.key).Result()

	return err

}

//检查锁是否被释放，未被释放就延长锁时间
func (lk *lock) checkLockIsRelease() {

	for {

		//需要提前一段时间验证锁是否还存在，否则锁会被redis释放
		checkCxt, c := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(lk.expiration.Milliseconds()-lk.expiration.Milliseconds()/5))

		select {

		//被手动释放锁
		case <-lk.cancelCxt.Done():

			//fmt.Println("手动释放1111")

			c()

			return

		case <-checkCxt.Done():

			//多次续期，直到锁被释放
			isContinue := lk.done()

			if !isContinue {

				c()

				return
			}

		}

	}
}

//判断锁是否已被释放，还未释放就续期
func (lk *lock) done() bool {

	lk.mu.Lock()

	select {

	case <-lk.cancelCxt.Done():

		return false

	default:

	}

	defer lk.mu.Unlock()

	cxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	res, err := lk.connect.Exists(cxt, lk.key).Result()

	cancel()

	if err != nil {

		return false
	}

	if res == 1 {

		cxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

		ok, err := lk.connect.Expire(cxt, lk.key, lk.expiration).Result()

		cancel()

		if err != nil {

			return false
		}

		if ok {

			//fmt.Println("续期")

			return true

		}

	}

	return false
}

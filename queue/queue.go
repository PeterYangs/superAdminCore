package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PeterYangs/superAdminCore/component/logs"
	"github.com/PeterYangs/superAdminCore/queue/register"
	"github.com/PeterYangs/superAdminCore/queue/template"
	"github.com/PeterYangs/superAdminCore/redis"
	"github.com/PeterYangs/tools"
	"github.com/PeterYangs/waitTree"
	redis2 "github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cast"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type job struct {
	Delay_  time.Duration `json:"-"` //延迟
	Data    template.Task `json:"data"`
	Queue_  string        `json:"queue"` //队列名称
	Id      string        `json:"id"`
	Time    string        `json:"time"`
	TryTime int           `json:"tryTime"` //重试次数
}

var once sync.Once

func init() {

	register.Handles.Init()

}

func Run(cxt context.Context, wait *waitTree.WaitTree) {

	defer func() {
		if r := recover(); r != nil {

			fmt.Println(r)

			fmt.Println(string(debug.Stack()))

			msg := fmt.Sprint(r)

			logs.NewLogs().Error(msg)

		}
	}()

	defer wait.Done()

	//延迟队列
	once.Do(func() {

		go checkDelay(cxt, wait)
	})

	queues := tools.Explode(",", os.Getenv("QUEUES"))

	for i, queue := range queues {

		queues[i] = os.Getenv("QUEUE_PREFIX") + queue
	}

	queueContext, queueCancel := context.WithCancel(context.Background())

	go func() {

		select {

		case <-cxt.Done():

			queueCancel()

			return

		}

	}()

	//错误等待时间
	sleepTime := 0

	//最大等待时间
	maxSleepTime := 1000 * 60

	for {

		select {

		case <-cxt.Done():

			fmt.Println("即时队列退出")

			return

		default:

		}

		//timeout为0则为永久超时
		s, err := redis.GetClient().BLPop(queueContext, 0, queues...).Result()

		if err != nil {

			if err.Error() != "context canceled" {

				fmt.Println(err)
			}

			//每错误一次就加大睡眠时间
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)

			if sleepTime < maxSleepTime {

				sleepTime += 100
			}

			continue
		}

		type q struct {
			Data    interface{} `json:"data"`
			Queue   string      `json:"queue"`
			Id      string      `json:"id"`
			Time    string      `json:"time"`
			TryTime int         `json:"tryTime"`
		}

		var qs q

		sleepTime = 0

		err = json.Unmarshal([]byte(s[1]), &qs)

		if err != nil {

			logs.NewLogs().Error(err.Error()).Stdout()

			continue
		}

		type d struct {
			TaskName   string      `json:"TaskName"`
			Parameters interface{} `json:"Parameters"`
		}

		var ds d

		err = mapstructure.Decode(qs.Data, &ds)

		if err != nil {

			logs.NewLogs().Error(err.Error()).Stdout()

			continue
		}

		//获取task
		hh, ok := register.Handles.GetTask(ds.TaskName)

		if !ok {

			fmt.Println("获取task失败")

			continue
		}

		//绑定参数
		hh.BindParameters(cast.ToStringMap(ds.Parameters))

		//执行任务
		runErr := hh.Run()

		if runErr != nil {

			logs.NewLogs().Error(runErr.Error()).Stdout()

			if qs.TryTime <= 0 {

				continue
			}

			//失败重试
			qs.TryTime--

			reTry, jsonErr := json.Marshal(qs)

			if jsonErr != nil {

				logs.NewLogs().Error(jsonErr.Error()).Stdout()

				continue
			}

			//头部插入，先执行
			redis.GetClient().LPush(context.TODO(), os.Getenv("QUEUE_PREFIX")+os.Getenv("DEFAULT_QUEUE"), reTry).Result()

		}

	}

}

func checkDelay(cxt context.Context, wait *waitTree.WaitTree) {

	defer func() {
		if r := recover(); r != nil {

			fmt.Println(r)

			fmt.Println(string(debug.Stack()))

			msg := fmt.Sprint(r)

			logs.NewLogs().Error(msg)

		}
	}()

	defer wait.Done()

	for {

		select {

		case <-cxt.Done():

			fmt.Println("延迟队列退出")

			return

		default:

		}

		push()

		time.Sleep(1 * time.Second)

	}

}

//延迟任务检查
func push() {

	//分布式锁
	lock := redis.GetClient().Lock("queue:delay:lock", 10*time.Second)

	defer lock.Release()

	if !lock.Get() {

		time.Sleep(1 * time.Second)

		return
	}

	//查询已到期任务
	list, err := redis.GetClient().ZRangeByScore(context.TODO(), os.Getenv("QUEUE_PREFIX")+"delay", &redis2.ZRangeBy{
		Min: "0",
		Max: cast.ToString(time.Now().Unix()),
	}).Result()

	if err != nil {

		fmt.Println(err)

		time.Sleep(1 * time.Second)
		return
	}

	for _, s := range list {

		var jsons map[string]interface{}

		json.Unmarshal([]byte(s), &jsons)

		queue := ""

		if jsons["queue"].(string) == "" {

			queue = os.Getenv("QUEUE_PREFIX") + os.Getenv("DEFAULT_QUEUE")

		} else {

			queue = os.Getenv("QUEUE_PREFIX") + jsons["queue"].(string)
		}

		//头部插入，先执行
		redis.GetClient().LPush(context.TODO(), queue, s).Result()

	}

	if len(list) > 0 {

		//删除已到期的任务
		redis.GetClient().ZRemRangeByRank(context.TODO(), os.Getenv("QUEUE_PREFIX")+"delay", 0, int64(len(list)-1))
	}

}

func Dispatch(task template.Task) *job {

	return &job{
		Data:   task,
		Delay_: 0,
		Id:     uuid.NewV4().String(),
	}

}

func (j *job) Delay(duration time.Duration) *job {

	j.Delay_ = duration

	if j.Delay_.Seconds() != 0 {

		j.Time = tools.Date("Y-m-d H:i:s", time.Now().Unix()+int64(j.Delay_.Seconds()))
	}

	return j
}

// SetTryTime 错误重试次数
func (j *job) SetTryTime(time int) *job {

	j.TryTime = time

	return j

}

func (j *job) Queue(queue string) *job {

	j.Queue_ = queue

	return j
}

func (j *job) Run() {

	queue := ""

	if j.Delay_ == 0 {

		if j.Queue_ == "" {

			queue = os.Getenv("QUEUE_PREFIX") + os.Getenv("DEFAULT_QUEUE")

			j.Queue_ = os.Getenv("DEFAULT_QUEUE")

		} else {

			queue = os.Getenv("QUEUE_PREFIX") + j.Queue_
		}

		data, err := json.Marshal(j)

		if err != nil {

			fmt.Println(err)

			return
		}

		redis.GetClient().RPush(context.TODO(), queue, data)

	} else {

		if j.Queue_ == "" {

			queue = os.Getenv("QUEUE_PREFIX") + "delay"

			j.Queue_ = os.Getenv("DEFAULT_QUEUE")

		} else {

			queue = os.Getenv("QUEUE_PREFIX") + "delay"
		}

		data, err := json.Marshal(j)

		if err != nil {

			fmt.Println(err)

			return
		}

		redis.GetClient().ZAdd(context.TODO(), queue, &redis2.Z{
			Score:  float64(time.Now().Unix() + cast.ToInt64(j.Delay_.Seconds())),
			Member: data,
		})
	}

}

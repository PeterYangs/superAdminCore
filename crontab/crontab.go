package crontab

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/v2/component/logs"
	"github.com/PeterYangs/waitTree"
)

type Crontab struct {
	schedules []*schedule
	quitWait  *waitTree.WaitTree
}

type schedule struct {

	//秒、分、小时、天、月、年，以秒换算
	//
	//year   int
	month   *number
	day     *number
	hour    *number
	minute  *number
	second  *number
	week    *number //0-6
	crontab *Crontab
	fn      func()
	//first   bool
}

type number struct {
	every   bool //每
	value   int  //数值
	between *between
}

type between struct {
	min int
	max int
}

func (c *Crontab) NewSchedule() *schedule {

	s := &schedule{
		crontab: c,
	}

	c.schedules = append(c.schedules, s)

	return s
}

// EveryDay 每天
func (s *schedule) EveryDay() *schedule {

	s.day = &number{
		every: true,
		value: 1,
	}

	return s

}

// DayAt 某天
func (s *schedule) DayAt(day int) *schedule {

	s.day = &number{

		value: day,
	}

	return s

}

// EveryDayAt 每几天
func (s *schedule) EveryDayAt(day int) *schedule {

	s.day = &number{
		value: day,
		every: true,
	}

	return s

}

// DayBetween 天，时间区间
func (s *schedule) DayBetween(min, max int) *schedule {

	s.day = &number{
		between: &between{
			min: min,
			max: max,
		},
	}

	return s

}

// EveryHour 每小时
func (s *schedule) EveryHour() *schedule {

	s.hour = &number{
		every: true,
		value: 1,
	}

	return s

}

// HourlyAt 某一个小时
func (s *schedule) HourlyAt(hour int) *schedule {

	s.hour = &number{
		value: hour,
	}

	return s
}

// EveryHourAt 每几个小时
func (s *schedule) EveryHourAt(hour int) *schedule {

	s.hour = &number{
		value: hour,
		every: true,
	}

	return s

}

// HourBetween 小时，时间区间
func (s *schedule) HourBetween(min, max int) *schedule {

	s.hour = &number{
		between: &between{
			min: min,
			max: max,
		},
	}

	return s

}

// EveryMinute 每分钟
func (s *schedule) EveryMinute() *schedule {

	s.minute = &number{
		value: 1,
		every: true,
	}

	return s
}

// EveryMinuteAt 每几分钟
func (s *schedule) EveryMinuteAt(minute int) *schedule {

	s.minute = &number{
		value: minute,
		every: true,
	}

	return s
}

// MinuteAt 某个分钟时间点
func (s *schedule) MinuteAt(minute int) *schedule {

	s.minute = &number{
		value: minute,
	}

	return s

}

func (s *schedule) Function(fun func()) {

	f := func() {

		//定时任务安全退出
		s.crontab.quitWait.Add(1)

		//捕获协程异常
		defer func() {

			if r := recover(); r != nil {

				msg := fmt.Sprint(r)

				msg = logs.NewLogs().Error(msg).Message()

				fmt.Println(msg)

			}

			s.crontab.quitWait.Done()

		}()

		fun()

	}

	s.fn = f
}

package crontab

import (
	"github.com/PeterYangs/waitTree"
	"github.com/spf13/cast"
	"time"
)

func Run(wait *waitTree.WaitTree, Registered func(*Crontab)) {

	_crontab := &Crontab{
		quitWait: wait,
	}

	//Registered(_crontab)

	Registered(_crontab)

	start := true

	delay := true

	//var diff time.Duration

	for {

		//准点校对
		if delay {

			for {

				if time.Now().Second() == 0 {

					delay = false

					break
				}

				//time.Sleep(1 * time.Second)

				time.Sleep(200 * time.Millisecond)

			}

		}

		if !start {

			//消除时间误差
			//time.Sleep(time.Minute * 1)

			time.Sleep(time.Duration((time.Duration((time.Minute*1).Seconds()-cast.ToFloat64(time.Now().Second())) * time.Second).Seconds()) * time.Second)

		}

		//startTime := time.Now()

		now := time.Now()

		go deal(_crontab, now)

		start = false

		//计算时间误差
		//diff = time.Now().Sub(startTime)

		//fmt.Println(diff)

	}

}

func deal(crontab *Crontab, now time.Time) {

	for _, s := range crontab.schedules {

		if s.day != nil {

			dealDay(s, now)

			continue
		}

		if s.hour != nil {

			dealHour(s, now)

			continue
		}

		if s.minute != nil {

			dealMinute(s, now)

			continue
		}

	}

}

func dealMinute(s *schedule, now time.Time) {

	if s.minute.every {

		if now.Minute()%s.minute.value == 0 {

			go s.fn()
		}

	} else {

		if now.Minute() == s.minute.value {

			go s.fn()
		}

	}

}

func dealHour(s *schedule, now time.Time) {

	if s.minute == nil {

		if now.Minute() == 0 {

			if s.hour.every {

				if now.Hour()%s.hour.value == 0 {

					go s.fn()
				}

			} else {

				//时间区间
				if s.hour.between != nil {

					if now.Hour() >= s.hour.between.min && now.Hour() <= s.hour.between.max {

						go s.fn()
					}

				} else {

					if now.Hour() == s.hour.value {

						go s.fn()
					}
				}

			}

		}

	} else {

		if s.hour.every {

			if now.Hour()%s.hour.value == 0 {

				//go s.fn()

				dealMinute(s, now)

			}

		} else {

			//时间区间
			if s.hour.between != nil {

				if now.Hour() >= s.hour.between.min && now.Hour() <= s.hour.between.max {

					//go s.fn()

					dealMinute(s, now)
				}

			} else {

				if now.Hour() == s.hour.value {

					//go s.fn()

					dealMinute(s, now)

				}
			}

		}

	}

}

func dealDay(s *schedule, now time.Time) {

	if s.hour == nil {

		if now.Hour() == 0 && now.Minute() == 0 {

			if s.day.every {

				if now.Day()%s.day.value == 0 {

					go s.fn()
				}

			} else {

				//时间区间
				if s.day.between != nil {

					if now.Day() >= s.day.between.min && now.Day() <= s.day.between.max {

						go s.fn()
					}

				} else {

					if now.Day() == s.day.value {

						go s.fn()
					}
				}

			}

		}

	} else {

		if s.day.every {

			if now.Day()%s.day.value == 0 {

				//go s.fn()

				dealHour(s, now)

			}

		} else {

			//时间区间
			if s.day.between != nil {

				if now.Day() >= s.day.between.min && now.Day() <= s.day.between.max {

					//go s.fn()

					dealHour(s, now)
				}

			} else {

				if now.Day() == s.day.value {

					//go s.fn()

					dealHour(s, now)

				}
			}

		}

	}

}

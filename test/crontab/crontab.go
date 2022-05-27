package crontab

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/v2/crontab"
)

func Crontab(crontab *crontab.Crontab) {

	//crontab.

	crontab.NewSchedule().EveryMinute().Function(func() {

		fmt.Println("每分钟")

	})

	crontab.NewSchedule().EveryMinuteAt(2).Function(func() {

		fmt.Println("每两分钟")
	})

}

package crontab

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/crontab"
)

func Crontab(crontab *crontab.Crontab) {

	//crontab.

	crontab.NewSchedule().EveryMinute().Function(func() {

		fmt.Println("每分钟")

	})

}

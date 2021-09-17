package demo

import "github.com/PeterYangs/superAdminCore/component/logs"

type Demo struct {
}

func (d Demo) GetName() string {

	return "demo"
}

func (d Demo) ArtisanRun() {

	logs.NewLogs().Debug("demo")
}

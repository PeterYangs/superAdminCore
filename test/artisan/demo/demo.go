package demo

import "github.com/PeterYangs/superAdminCore/v2/component/logs"

type Demo struct {
}

func (d Demo) GetName() string {

	return "自定义命令"
}

func (d Demo) ArtisanRun() {

	logs.NewLogger().Error("123")
	logs.NewLogger().Info("123")

}

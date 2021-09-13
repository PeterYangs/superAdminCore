package artisan

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/artisan/migrates"
	"github.com/manifoldco/promptui"
)

type artisan interface {
	ArtisanRun()
}

func Artisan() {

	prompt := promptui.Select{
		Label: "选择类型",
		Items: []string{"数据库迁移", "数据填充", "生成key", "生成任务类"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	switch result {

	case "数据库迁移":

		new(migrates.MigrateRun).ArtisanRun()

		//case "数据填充":
		//
		//	new(bin2.Bin).Run()
		//
		//case "生成key":
		//
		//	new(key.Key).Run()
		//
		//case "生成任务类":
		//
		//	new(task.TaskCmd).Run()

	}

}

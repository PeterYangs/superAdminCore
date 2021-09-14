package artisan

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/artisan/migrates"
	"github.com/PeterYangs/superAdminCore/artisan/queue"
	"github.com/manifoldco/promptui"
	"log"
)

type Artisan interface {
	ArtisanRun()
	GetName() string
}

func RunArtisan(artisan ...Artisan) {

	//内置命令
	list := []Artisan{
		new(migrates.MigrateRun),
		new(queue.QueueRun),
	}

	//加载自定义命令
	list = append(list, artisan...)

	var nameList []string
	nameMap := make(map[string]Artisan)

	for _, a := range list {

		nameList = append(nameList, a.GetName())
		nameMap[a.GetName()] = a

	}

	prompt := promptui.Select{
		Label: "选择类型",
		Items: nameList,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	a, ok := nameMap[result]

	if !ok {

		log.Println("命令不存在")

		return
	}

	a.ArtisanRun()

}

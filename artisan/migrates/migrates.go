package migrates

import (
	"context"
	"fmt"
	"github.com/PeterYangs/gcmd2"
	"github.com/PeterYangs/superAdminCore/database"
	"github.com/PeterYangs/superAdminCore/mod"
	"github.com/PeterYangs/superAdminCore/model"
	"github.com/PeterYangs/tools"
	"github.com/manifoldco/promptui"
	"io/ioutil"
	"os"
	"time"
)

type MigrateRun struct {
}

func (m MigrateRun) GetName() string {

	return "数据库迁移"
}

func (m MigrateRun) ArtisanRun() {

	prompt := promptui.Select{
		Label: "选择类型",
		Items: []string{"创建数据库迁移", "执行迁移", "回滚迁移"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	if result == "创建数据库迁移" {

		prompt := promptui.Select{
			Label: "选择类型",
			Items: []string{"创建表", "修改表"},
		}

		_, result, err = prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		if result == "创建表" {

			//CreateMigration()
			prompt := promptui.Prompt{
				Label: "输入表名",
				//Validate: validate,
			}

			result, err := prompt.Run()

			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			CreateMigration(result, "create")

		}

		if result == "修改表" {

			//CreateMigration()
			prompt := promptui.Prompt{
				Label: "输入表名",
				//Validate: validate,
			}

			result, err := prompt.Run()

			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			CreateMigration(result, "update")

		}

	}

	if result == "执行迁移" {

		prompt := promptui.Select{
			Label: "确定吗？",
			Items: []string{"是", "否"},
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		if result == "是" {

			Up()
		}

	}

	if result == "回滚迁移" {

		prompt := promptui.Select{
			Label: "确定吗？",
			Items: []string{"是", "否"},
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		if result == "是" {

			Down()
		}

	}

}

// CreateMigration 生成数据库迁移文件
func CreateMigration(table string, types string) {

	name := "migrate_" + tools.Date("Y_m_d_His", time.Now().Unix()) + "_" + types + "_" + table + "_table"

	os.Mkdir("./migrate/migrations/"+name, 0755)

	if types == "create" {

		CreateTable(name, table)
	}

	if types == "update" {

		UpdateTable(name, table)
	}

}

func UpdateTable(pack string, table string) {

	//生成迁移文件夹
	os.MkdirAll("migrations/"+pack, 0755)

	f, err := os.OpenFile("migrations/"+pack+"/migrate.go", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)

	if err != nil {

		panic(err)

	}

	defer f.Close()

	f.Write([]byte(`package ` + pack + `

import "github.com/PeterYangs/superAdminCore/migrate"

func Up() {

	migrate.Table("` + table + `", func(createMigrate *migrate.Migrate) {

		createMigrate.Name = "` + pack + `"


	})

}

func Down() {
	
	migrate.Table("` + table + `", func(createMigrate *migrate.Migrate) {

		

	})


}



`))

}

func CreateTable(pack string, table string) {

	//生成迁移文件夹
	os.MkdirAll("migrations/"+pack, 0755)

	f, err := os.OpenFile("migrations/"+pack+"/migrate.go", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)

	if err != nil {

		panic(err)

	}

	defer f.Close()

	f.Write([]byte(`package ` + pack + `

import "github.com/PeterYangs/superAdminCore/migrate"

func Up() {

	migrate.Create("` + table + `", func(createMigrate *migrate.Migrate) {

		createMigrate.Name = "` + pack + `"

		createMigrate.BigIncrements("id")
		
		createMigrate.Timestamp("created_at").Nullable()

		createMigrate.Timestamp("updated_at").Nullable()

		// createMigrate.Timestamp("deleted_at").Nullable()
		

	})

}

func Down() {

	migrate.DropIfExists("` + table + `")

}



`))

}

// Up 执行迁移
func Up() {

	fileInfo, _ := ioutil.ReadDir("migrations")

	//for _, info := range fileInfo {

	os.MkdirAll("migrations/bin", 0755)

	f, _ := os.OpenFile("migrations/bin/X.go", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)

	f.Write([]byte(`package main

import (
 ` + getPackageList(fileInfo) + `

"github.com/PeterYangs/superAdminCore/migrate/transaction"

)

//func init() {
//
//	//加载配置文件
//	err := godotenv.Load("./.env")
//	if err != nil {
//		panic("配置文件加载失败")
//	}
//
//}


func main() {


	transaction.E = nil

   ` + getFuncList(fileInfo) + `

}
`))

	//}

	//gcmd.Command("go run migrate/bin/X.go").Start()

	err := gcmd2.NewCommand("go run migrations/bin/X.go", context.TODO()).SetSystemEnv().Start()

	if err != nil {

		fmt.Println(err)
	}

}

func getPackageList(f []os.FileInfo) string {

	str := ""

	for _, info := range f {

		if info.Name() == ".gitignore" || info.Name() == "bin" {

			continue
		}

		str += "\"" + mod.GetModuleName() + "/migrations/" + info.Name() + "\"\n"

	}

	return str

}

func getFuncList(f []os.FileInfo) string {

	str := ""

	for _, info := range f {

		if info.Name() == ".gitignore" || info.Name() == "bin" {

			continue
		}

		str += info.Name() + ".Up()" + "\n"

	}

	return str

}

// Down 迁移回滚
func Down() {

	//database.GetDb().
	var migration model.Migrations

	re := database.GetDb().Order("id desc").First(&migration)

	if re.Error != nil {

		return
	}

	batch := migration.Batch

	migrations := make([]*model.Migrations, 0)

	database.GetDb().Model(&model.Migrations{}).Where("batch = ?", batch).Find(&migrations)

	f, _ := os.OpenFile("migrations/bin/Y.go", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)

	f.Write([]byte(`package main

import (
 ` + getPackageListForDown(migrations) + `
"github.com/PeterYangs/superAdminCore/migrate/transaction"
)




func main() {


	transaction.E = nil

   ` + getFuncListForDown(migrations) + `

}
`))

	//gcmd.Command("go run migrate/bin/Y.go").Start()

	err := gcmd2.NewCommand("go run migrations/bin/Y.go", context.TODO()).SetSystemEnv().Start()

	if err != nil {

		fmt.Println(err)

		return
	}

	for _, m := range migrations {

		database.GetDb().Delete(m)

	}

}

func getPackageListForDown(m []*model.Migrations) string {

	str := ""

	for _, info := range m {

		//str += "\"gin-web/migrate/migrations/" + info.Migration + "\"\n"

		str += "\"" + mod.GetModuleName() + "/migrations/" + info.Migration + "\"\n"

	}

	return str

}

func getFuncListForDown(m []*model.Migrations) string {

	str := ""

	for _, info := range m {

		str += info.Migration + ".Down()" + "\n"

	}

	return str

}

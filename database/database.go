package database

import (
	"fmt"
	"github.com/spf13/cast"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"os"
	"sync"
	"time"
)

var db *gorm.DB

var mysqlOnce sync.Once

func connect() {

	var err error

	dsn := os.Getenv("DB_USERNAME") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_DATABASE") + "?charset=utf8mb4&parseTime=True&loc=Local"

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {

		fmt.Println(err)

	}

	s, _ := db.DB()

	//连接池打开最大连接数
	s.SetMaxOpenConns(cast.ToInt(os.Getenv("DB_MAX_OPEN_CONNS")))

	//连接池最大空闲连接数
	s.SetMaxIdleConns(cast.ToInt(os.Getenv("DB_MAX_IDLE_CONNS")))

	//设置连接过期时间
	s.SetConnMaxLifetime(1 * time.Minute)

}

func GetDb() *gorm.DB {

	//数据库连接单例
	mysqlOnce.Do(func() {

		connect()

	})

	return db

}

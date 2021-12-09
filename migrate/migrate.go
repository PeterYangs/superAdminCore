package migrate

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/database"
	"github.com/PeterYangs/superAdminCore/migrate/transaction"
	"github.com/PeterYangs/superAdminCore/model"
	"github.com/PeterYangs/tools"
	"github.com/spf13/cast"
	"strings"
	"sync"
)

var batch = 1

var once sync.Once

type NullValue int

const Null NullValue = 0x00000

type Tag int

const (
	CREATE Tag = 0x00000
	UPDATE Tag = 0x00001
	DELETE Tag = 0x00002
)

type Types string

const (
	Int       Types = "int"
	Bigint    Types = "bigint"
	String    Types = "varchar"
	Text      Types = "text"
	Timestamp Types = "timestamp"
	Enum      Types = "enum"
)

func (t Types) ToString() string {

	return string(t)
}

type Migrate struct {
	Tag    Tag
	Table  string
	fields []*field
	Name   string
	unique [][]string //[ [name,title]  ]
}

type field struct {
	column       string //字段名称
	isPrimaryKey bool   //主键
	isUnsigned   bool   //无符号
	isNullable   bool
	types        Types //数据类型
	length       int   //长度
	tag          Tag
	defaultValue interface{}
	comment      string
	unique       bool     //唯一索引
	enumList     []string //枚举列表
}

func getBatch() {

	once.Do(func() {

		var migrations model.Migrations

		re := database.GetDb().Order("id desc").First(&migrations)

		if re.Error == nil {

			batch = migrations.Batch + 1

		}

	})

}

// Create 创建表
func Create(table string, callback func(*Migrate)) {

	getBatch()

	m := &Migrate{
		Table: table,
		Tag:   CREATE,
	}

	defer func() {

		run(m)

	}()

	callback(m)

}

func Table(table string, callback func(*Migrate)) {

	getBatch()

	m := &Migrate{
		Table: table,
		Tag:   UPDATE,
	}

	defer func() {

		run(m)

	}()

	callback(m)

}

func DropIfExists(table string) {

	database.GetDb().Exec("drop table if exists `" + table + "`")

}

// BigIncrements 主键字段
func (c *Migrate) BigIncrements(column string) {

	c.fields = append(c.fields, &field{column: column, isPrimaryKey: true})
}

// Unique 设置唯一索引
func (c *Migrate) Unique(column ...string) {

	c.unique = append(c.unique, column)

}

// Integer int
func (c *Migrate) Integer(column string) *field {

	f := &field{column: column, types: Int, length: 10, tag: CREATE}

	c.fields = append(c.fields, f)

	return f
}

// BigInteger bigint
func (c *Migrate) BigInteger(column string) *field {

	f := &field{column: column, types: Bigint, length: 15, tag: CREATE}

	c.fields = append(c.fields, f)

	return f

}

func (c *Migrate) String(column string, length int) *field {

	f := &field{column: column, types: String, length: length, tag: CREATE}

	c.fields = append(c.fields, f)

	return f

}

func (c *Migrate) Enum(column string, allowed []string) *field {

	f := &field{column: column, types: Enum, tag: CREATE, enumList: allowed}

	c.fields = append(c.fields, f)

	return f

}

func (c *Migrate) Text(column string) *field {

	f := &field{column: column, types: Text, tag: CREATE}

	c.fields = append(c.fields, f)

	return f
}

func (c *Migrate) Timestamp(column string) *field {

	f := &field{column: column, types: Timestamp, tag: CREATE}

	c.fields = append(c.fields, f)

	return f
}

// DropColumn 删除字段
func (c *Migrate) DropColumn(column string) {

	f := &field{column: column, tag: DELETE}

	c.fields = append(c.fields, f)

}

func (f *field) Default(value interface{}) *field {

	f.defaultValue = value

	return f
}

func (f *field) Comment(comment string) *field {

	f.comment = comment

	return f

}

func (f *field) Change() {

	//f.isChange = true
	f.tag = UPDATE

}

// Unsigned 无符号
func (f *field) Unsigned() *field {

	f.isUnsigned = true

	return f
}

// Unique 唯一索引
func (f *field) Unique() *field {

	f.unique = true

	return f
}

func (f *field) Nullable() *field {

	f.isNullable = true

	return f
}

func run(m *Migrate) {

	if transaction.E != nil {

		return
	}

	checkMigrationsTable()

	isFind := database.GetDb().Where("migration = ?", m.Name).First(&model.Migrations{})

	//已存在的迁移不执行
	if isFind.Error == nil {

		fmt.Println("find:" + m.Name)

		return
	}

	//batch := 1

	if m.Tag == CREATE {

		sql := "CREATE TABLE `" + m.Table + "` (" +
			"`" + getPrimaryKey(m) + "` int(10) unsigned NOT NULL AUTO_INCREMENT," +
			setTableUnique(m) +
			getColumn(m) +
			setColumnUnique(m) +
			"PRIMARY KEY (`" + getPrimaryKey(m) + "`)" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"

		t := database.GetDb().Exec(sql)

		if t.Error != nil {

			fmt.Println(sql)

			transaction.E = t.Error

			return
		}

		database.GetDb().Create(&model.Migrations{
			Migration: m.Name,
			Batch:     batch,
		})

	}

	if m.Tag == UPDATE {

		sql := "alter table `" + m.Table + "` "

		for i, f := range m.fields {

			switch f.tag {

			case CREATE:

				sql += " add column  " + setColumnAttr(f)

			case UPDATE:

				sql += " modify " + setColumnAttr(f)

			case DELETE:

				sql += " drop COLUMN `" + f.column + "` "

			}

			if i+1 < len(m.fields) {

				sql += ","
			}

		}

		//索引添加
		for i, strings := range m.unique {

			if len(m.fields) > 0 {

				sql += ","
			}

			sql += " add unique  `" + tools.Join("+", strings) + "` (`" + tools.Join("`,`", strings) + "`)" + " USING BTREE"

			if i+1 < len(m.unique) {

				sql += ","
			}

		}

		t := database.GetDb().Exec(sql)

		if t.Error != nil {

			fmt.Println(t.Error)

			transaction.E = t.Error

			return
		}

		if m.Name != "" {

			//添加迁移记录
			database.GetDb().Create(&model.Migrations{
				Migration: m.Name,
				Batch:     batch,
			})

		}

	}

}

func getPrimaryKey(m *Migrate) string {

	id := ""

	for _, f := range m.fields {

		if f.isPrimaryKey {

			id = f.column
		}
	}

	if id == "" {

		panic("主键不能为空")
	}

	return id
}

func getColumn(m *Migrate) string {

	str := ""

	for _, f := range m.fields {

		if f.isPrimaryKey {

			continue

		}

		str += setColumnAttr(f)

		str += ","

	}

	return str
}

//设置字段唯一索引
func setColumnUnique(m *Migrate) string {

	str := ""

	for _, f := range m.fields {

		if f.unique {

			str += " UNIQUE KEY `" + f.column + "` (`" + f.column + "`), "

		}

	}

	return str

}

func setTableUnique(m *Migrate) string {

	str := ""

	for _, strings := range m.unique {

		str += " UNIQUE KEY `" + tools.Join("+", strings) + "` (`" + tools.Join("`,`", strings) + "`)" + " USING BTREE, "

	}

	return str
}

//设置字段类型
func setColumnAttr(f *field) string {

	str := ""

	switch f.types {

	case Text:

		str += " `" + f.column + "` " + f.types.ToString() + " "

		break

	case Timestamp:

		str += " `" + f.column + "` " + f.types.ToString() + " NULL "

		break

	case Enum:

		enumTemp := make([]string, len(f.enumList))

		for i, s := range f.enumList {

			enumTemp[i] = `'` + strings.Replace(s, `'`, `\'`, -1) + `'`
		}

		str += " `" + f.column + "` " + "enum(" + tools.Join(`,`, enumTemp) + ")" + " "

		break

	default:

		str += " `" + f.column + "` " + f.types.ToString() + "(" + cast.ToString(f.length) + ") "

	}

	if f.isUnsigned {

		str += " unsigned "
	}

	if !f.isNullable && f.defaultValue != Null {

		str += " NOT NULL "
	}

	switch f.defaultValue.(type) {

	case NullValue:

		str += " DEFAULT NULL "

		break

	case string:

		str += " DEFAULT '" + cast.ToString(f.defaultValue) + "' "

	case int:

		str += " DEFAULT '" + cast.ToString(f.defaultValue) + "' "

	}

	if f.comment != "" {

		str += " COMMENT '" + f.comment + "' "
	}

	return str
}

// CheckMigrationsTable 检查数据迁移表是否存在
func checkMigrationsTable() {

	database.GetDb().Exec("CREATE TABLE IF NOT EXISTS `migrations` (`id` int(10) unsigned NOT NULL AUTO_INCREMENT,  `migration` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,  `batch` int(11) NOT NULL,  PRIMARY KEY (`id`)) ENGINE=InnoDB AUTO_INCREMENT=63 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")

}

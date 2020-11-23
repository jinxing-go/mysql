package mysql

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var id int32 = 0

func GetUniqueName(prefix string) string {
	atomic.AddInt32(&id, 1)
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), id)
}

func GetEnv(name, defaultValue string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	}

	return defaultValue
}

func getDsn(name string) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:3306)",
		GetEnv("TEST_DB_USERNAME", "root"),
		GetEnv("TEST_DB_PASSWORD", ""),
		GetEnv("TEST_DB_HOST", "127.0.0.1"),
	) + "/" + name + "?charset=utf8&parseTime=True&loc=Asia%2FShanghai"
}

func NewTestDatabase(t *testing.T, schema string, fixtures ...string) *MySQl {
	database := GetUniqueName("test")
	mySQL := NewMySQL(&MySQLConfig{
		Dsn:         getDsn(""),
		Driver:      "mysql",
		ShowSql:     false,
		MaxLifetime: 30,
	})

	name := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database)
	fmt.Println("- 🛠️ " + name)
	_, err := mySQL.Exec(name)
	if err != nil {
		panic("Create database error: " + err.Error())
	}

	mySQL = NewMySQL(&MySQLConfig{
		Dsn:         getDsn(database),
		Driver:      "mysql",
		ShowSql:     false,
		MaxLifetime: 30,
	})

	// 读取数据库表结构文件
	fmt.Println("- ‍🚀 Import schema: " + schema)
	str, err := ioutil.ReadFile(schema)
	if err != nil {
		panic(err)
	}

	mySQL.Exec(string(str))

	for _, filename := range fixtures {
		if filename != "" {
			fmt.Println("- ‍🚀 Load fixtures: " + filename)
			str, err = ioutil.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			data := strings.Split(string(str), ";\n")
			for _, execSql := range data {
				mySQL.Exec(strings.TrimRight(execSql, ";"))
			}
		}
	}

	t.Cleanup(func() {
		mySQL.ShowSql = false
		dropName := fmt.Sprintf("DROP DATABASE IF EXISTS %s", database)
		fmt.Println("- 🗑️ " + dropName)
		mySQL.Exec(dropName)
	})

	mySQL.ShowSql = true
	return mySQL
}

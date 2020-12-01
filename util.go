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

// InStringSlice 验证字符串是否在slice里面
func InStringSlice(strSlice []string, need string) bool {
	if len(strSlice) == 0 {
		return false
	}

	for _, v := range strSlice {
		if v == need {
			return true
		}
	}

	return false
}

// Studly user_id to UserId
func Studly(key string) string {
	// user_id to `user id`
	s := strings.Replace(key, "_", " ", -1)
	// `user id` to `User Id`
	s = strings.Title(s)
	// `User Id` to `UserId`
	return strings.Replace(s, " ", "", -1)
}

// SliceToMap 将slice转为map
func SliceToMap(str []string) map[string]bool {
	data := make(map[string]bool)
	for _, key := range str {
		data[key] = true
	}

	return data
}

func ToQueryWhere(data interface{}, exceptColumn, joinColumn []string) ([]string, []interface{}) {
	where := make([]string, 0)
	args := make([]interface{}, 0)

	// 处理需要加入和排除的字段
	except, join := SliceToMap(exceptColumn), SliceToMap(joinColumn)

	columns := StructColumns(data, "db")
	for _, v := range columns {
		// 先排除
		if _, isExcept := except[v.Name]; isExcept {
			continue
		}

		// 需要加入
		_, isJoin := join[v.Name]
		if !v.IsZero || isJoin {
			where = append(where, fmt.Sprintf("`%s` = ?", v.Name))
			args = append(args, v.Value)
		}
	}

	return where, args
}

func getDsn(name string) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:3306)",
		GetEnv("TEST_DB_USERNAME", "root"),
		GetEnv("TEST_DB_PASSWORD", ""),
		GetEnv("TEST_DB_HOST", "127.0.0.1"),
	) + "/" + name + "?charset=utf8&parseTime=True&loc=Asia%2FShanghai"
}

func NewTestMySQL(t *testing.T, schema string, fixtures ...string) *MySQl {

	mySQL := NewMySQL(&Config{
		Dsn:         getDsn(""),
		Driver:      "mysql",
		ShowSql:     false,
		MaxLifetime: 30,
	})

	database := GetUniqueName("test")
	name := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database)
	fmt.Println("- ✅ " + name)
	mySQL.Exec(name)
	mySQL.Close()

	mySQL = NewMySQL(&Config{
		Dsn:         getDsn(database),
		Driver:      "mysql",
		ShowSql:     false,
		MaxLifetime: 30,
	})

	t.Cleanup(func() {
		mySQL.ShowSql(false)
		dropName := fmt.Sprintf("DROP DATABASE IF EXISTS %s", database)
		fmt.Println("- ✅ " + dropName)
		mySQL.Exec(dropName)
		mySQL.Close()
	})

	// 读取数据库表结构文件
	fmt.Println("- ✅ Import schema: " + schema)
	str, err := ioutil.ReadFile(schema)
	if err != nil {
		panic(err)
	}

	mySQL.Exec(string(str))

	for _, filename := range fixtures {
		if filename != "" {
			fmt.Println("- ✅ Load fixtures: " + filename)
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

	mySQL.ShowSql(true)
	return mySQL
}

func RunTestMySQL(t *testing.T, schema string, run func(mySQL *MySQl), fixtures ...string) {
	mySQL := NewTestMySQL(t, schema, fixtures...)
	run(mySQL)
}

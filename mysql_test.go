package mysql

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	examplePathName = "testdata/example.sql"
	userPathName    = "testdata/user.sql"
)

func TestNewMySQL(t *testing.T) {
	assert.Panics(t, func() {
		NewMySQL(&MySQLConfig{
			Dsn:     "root:test123456789@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=True&loc=Asia%2FShanghai",
			Driver:  "mysql",
			ShowSql: true,
		})
	})

	t.Run("正常连接", func(t *testing.T) {
		NewMySQL(&MySQLConfig{
			Dsn:     getDsn(""),
			Driver:  "mysql",
			ShowSql: true,
		})
	})
}

func TestMySQl_Find(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName, userPathName)
	// 执行正常
	err := mySQL.Find(&User{UserId: 1, Username: "test1"})
	assert.NoError(t, err)

	// 执行失败
	err1 := mySQL.Find(&User{UserId: 1001, Status: 1})
	assert.Error(t, err1)
	assert.Equal(t, sql.ErrNoRows, err1)
}

func TestMySQl_Delete(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName, userPathName)
	// 执行正常(删除一条)
	row1, err1 := mySQL.Delete(&User{UserId: 1})
	assert.NoError(t, err1)
	assert.Equal(t, int64(1), row1)

	// 执行正常(删除0条)
	row1, err1 = mySQL.Delete(&User{UserId: 10000})
	assert.NoError(t, err1)
	assert.Equal(t, int64(0), row1)
}

func TestMySQl_Update(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName, userPathName)
	// 执行正常：更新1条
	row, err := mySQL.Update(&User{UserId: 1, Username: "jinxing.liu123"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), row)

	// 执行正常：更新0条
	row, err = mySQL.Update(&User{UserId: 1001, Username: "jinxing.liu123"}, "status")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), row)
}

func TestMySQl_Create(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName)
	// 执行正常
	i := &User{
		Username: "my-name",
		Password: "123456",
	}

	err := mySQL.Create(i)
	assert.NoError(t, err)
	assert.NotEqual(t, int64(0), i.UserId)

	err1 := mySQL.Create(i)
	assert.Error(t, err1)
}

func TestMySQl_Exec(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName)
	// 执行正常删除
	row, err := mySQL.Exec("delete from `user` where `user_id` = ?", 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), row)

	// 执行错误删除
	row1, err1 := mySQL.Exec("delete from `users` where `user_id` = ?", 100)
	assert.Error(t, err1)
	assert.Equal(t, int64(0), row1)

	// 执行正常修改
	row2, err2 := mySQL.Exec("update `user` set `username` = ? where `user_id` = ?", "jinxing.liu", 100)
	assert.NoError(t, err2)
	assert.Equal(t, int64(0), row2)

	// 执行错误修改
	row3, err3 := mySQL.Exec("update `users` set `username` = ? where `user_id` = ?", "jinxing.liu", 100)
	assert.Error(t, err3)
	assert.Equal(t, int64(0), row3)
}

func TestMySQl_FindAll(t *testing.T) {
	users := make([]*User, 0)
	mySQL := NewTestMySQL(t, examplePathName, userPathName)
	err := mySQL.FindAll(&users, "status = ?", 1)
	assert.NoError(t, err)
	for _, v := range users {
		assert.Equal(t, true, v.UserId > 0)
	}

	assert.Panics(t, func() {
		mySQL.FindAll(nil, "")
	})
}

func TestMySQl_Builder(t *testing.T) {
	my := &MySQl{}
	assert.Equal(t, my.Builder(nil), &Builder{db: my, data: nil})
}

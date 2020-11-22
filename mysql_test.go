package mysql

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func getDsn() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/test?charset=utf8&parseTime=True&loc=Asia/Shanghai",
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_HOST"),
	)
}

func TestNewMySQL(t *testing.T) {
	assert.Panics(t, func() {
		NewMySQL(&MySQLConfig{
			Dsn:     "root:test123456789@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=True&loc=Asia%2FShanghai",
			Driver:  "mysql",
			ShowSql: true,
		})
	})
}

func newTestMySQL() *MySQl {
	return NewMySQL(&MySQLConfig{
		Dsn:         getDsn(),
		Driver:      "mysql",
		ShowSql:     true,
		MaxLifetime: 30,
	})
}

func TestMySQl_getPkValue(t *testing.T) {
	v := &MySQl{}
	s := v.getPkValue(&User{}, "UserId")
	assert.Equal(t, int64(0), s)
	assert.Equal(t, int64(1), v.getPkValue(&User{UserId: 1}, "user_id"))
}

func TestMySQl_toStudly(t *testing.T) {
	mySQL := &MySQl{}
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "test",
			args: "test",
			want: "Test",
		},
		{
			name: "test_name",
			args: "test_name",
			want: "TestName",
		},
		{
			name: "test name",
			args: "test name",
			want: "TestName",
		},
		{
			name: "TestName",
			args: "testName",
			want: "TestName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mySQL.toStudly(tt.args); got != tt.want {
				t.Errorf("toStudly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMySQl_Find(t *testing.T) {
	mySQL := newTestMySQL()
	// 执行正常
	err := mySQL.Find(&User{UserId: 1, Username: "jinxing.liu"})
	assert.NoError(t, err)

	// 执行失败
	err1 := mySQL.Find(&User{UserId: 1001, Status: 1})
	assert.Error(t, err1)
	assert.Equal(t, sql.ErrNoRows, err1)
}

func TestMySQl_Delete(t *testing.T) {
	// 执行正常
	row1, err1 := newTestMySQL().Delete(&User{UserId: 1001})
	assert.NoError(t, err1)
	assert.Equal(t, int64(0), row1)
}

func TestMySQl_Update(t *testing.T) {
	mySQL := newTestMySQL()
	// 执行正常
	row, err := mySQL.Update(&User{UserId: 1001, Username: "jinxing.liu123"})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), row)

	// 修改更新0值
	row, err = mySQL.Update(&User{UserId: 1001, Username: "jinxing.liu123"}, "status")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), row)
}

func TestMySQl_Create(t *testing.T) {
	mySQL := newTestMySQL()
	mySQL.Delete(&User{Username: "my-name"})
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
	mySQL := newTestMySQL()

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

func TestMySQl_AnalyticStructure(t *testing.T) {
	user := newTestMySQL().AnalyticStructure(&User{UserId: 1}, "")
	assert.Equal(t, 9, len(user))
}

func TestMySQl_toQueryWhere(t *testing.T) {
	db := newTestMySQL()
	// 没有排除、没有添加
	where, args := db.toQueryWhere(&User{
		UserId:   1,
		Username: "jinxing.liu",
	}, "", nil, nil)

	assert.Equal(t, 2, len(where))
	assert.Equal(t, 2, len(args))
	fmt.Println(where, args)

	// 有排除、无添加
	where, args = db.toQueryWhere(&User{
		UserId:   1,
		Username: "jinxing.liu",
	}, "", []string{"user_id"}, nil)

	assert.Equal(t, 1, len(where))
	assert.Equal(t, 1, len(args))
	fmt.Println(where, args)

	// 无排除、有添加
	where, args = db.toQueryWhere(&User{
		UserId:   1,
		Username: "jinxing.liu",
	}, "", nil, []string{"password", "status"})

	assert.Equal(t, 4, len(where))
	assert.Equal(t, 4, len(args))
	fmt.Println(where, args)

	// 有排除、有添加
	where, args = db.toQueryWhere(&User{
		UserId:   1,
		Username: "jinxing.liu",
	}, "", []string{"user_id"}, []string{"password", "status"})

	assert.Equal(t, 3, len(where))
	assert.Equal(t, 3, len(args))
	fmt.Println(where, args)
}

func TestMySQl_InStringSlice(t *testing.T) {
	mySQL := newTestMySQL()
	type args struct {
		strSlice []string
		need     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "测试正常-true",
			args: args{strSlice: []string{"username", "age", "status", "created_at"}, need: "age"},
			want: true,
		},
		{
			name: "测试正常-false",
			args: args{strSlice: []string{"username", "age", "status", "created_at"}, need: "age1"},
			want: false,
		},
		{
			name: "测试正常-false",
			args: args{strSlice: []string{}, need: "age1"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mySQL.InStringSlice(tt.args.strSlice, tt.args.need); got != tt.want {
				t.Errorf("InStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

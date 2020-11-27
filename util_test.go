package mysql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUniqueName(t *testing.T) {
	str := GetUniqueName("test")
	fmt.Println(str)
	assert.NotEqual(t, "", str)
}

func TestGetEnv(t *testing.T) {
	str := GetEnv("USER_NAME", "username")
	assert.Equal(t, "username", str)
	str = GetEnv("TEST_DB_USERNAME123", "")
	assert.Equal(t, "", str)
}

func TestNewTestMySQL(t *testing.T) {
	mySQL := NewTestMySQL(t, "testdata/example.sql")
	_, err := mySQL.Exec("SHOW TABLES")
	assert.NoError(t, err)
}

func TestNewTestMySQL_FixturesPanic(t *testing.T) {
	assert.Panics(t, func() {
		NewTestMySQL(t, "testdata/example.sql", "test.sql")
	})
}

func TestNewTestMySQL_SchemaPanic(t *testing.T) {
	assert.Panics(t, func() {
		NewTestMySQL(t, "testdata/example123.sql")
	})
}

func TestRunTestMySQL(t *testing.T) {
	RunTestMySQL(t, "testdata/example.sql", func(mySQL *MySQl) {
		_, err := mySQL.Exec("SHOW TABLES")
		assert.NoError(t, err)
	})
}

func TestInStringSlice(t *testing.T) {
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
			if got := InStringSlice(tt.args.strSlice, tt.args.need); got != tt.want {
				t.Errorf("InStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStudly(t *testing.T) {
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
			if got := Studly(tt.args); got != tt.want {
				t.Errorf("toStudly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSliceToMap(t *testing.T) {
	str := []string{"username", "age"}
	m := SliceToMap(str)
	assert.Equal(t, true, m["username"])
	assert.Equal(t, true, m["age"])
}

func TestToQueryWhere(t *testing.T) {
	s := struct {
		Username string `db:"username"`
		Age      int    `db:"age"`
	}{
		Username: "123",
		Age:      1,
	}

	where, args := ToQueryWhere(&s, nil, nil)
	assert.Equal(t, []string{"`username` = ?", "`age` = ?"}, where)
	assert.Equal(t, 2, len(args))
}

func TestToQueryWhere1(t *testing.T) {
	// 没有排除、没有添加
	where, args := ToQueryWhere(&User{
		UserId:   1,
		Username: "jinxing.liu",
	}, nil, nil)

	assert.Equal(t, 2, len(where))
	assert.Equal(t, 2, len(args))
	fmt.Println(where, args)

	// 有排除、无添加
	where, args = ToQueryWhere(&User{
		UserId:   1,
		Username: "jinxing.liu",
	}, []string{"user_id"}, nil)

	assert.Equal(t, 1, len(where))
	assert.Equal(t, 1, len(args))
	fmt.Println(where, args)

	// 无排除、有添加
	where, args = ToQueryWhere(&User{
		UserId:   1,
		Username: "jinxing.liu",
	}, nil, []string{"password", "status"})

	assert.Equal(t, 4, len(where))
	assert.Equal(t, 4, len(args))
	fmt.Println(where, args)

	// 有排除、有添加
	where, args = ToQueryWhere(&User{
		UserId:   1,
		Username: "jinxing.liu",
	}, []string{"user_id"}, []string{"password", "status"})

	assert.Equal(t, 3, len(where))
	assert.Equal(t, 3, len(args))
	fmt.Println(where, args)
}

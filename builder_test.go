package mysql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_All(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName, userPathName)
	user := make([]*User, 0)
	err := NewBuilder(mySQL, &user).Where("status", "in", 1, 2).All()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(user))

	err = NewBuilder(mySQL, &user).Where("status", "in", []interface{}{}).All()
	assert.Error(t, err)
}

func TestBuilder_One(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName, userPathName)
	user := &User{}
	err := NewBuilder(mySQL, user).Where("status", "in", 1, 2).One()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.UserId)

	err = NewBuilder(mySQL, user).Where("status", "in", []interface{}{}).One()
	fmt.Printf("err = %v \n", err)
	assert.Error(t, err)

	st := struct {
		Username string `db:"username"`
		Password string `db:"password"`
	}{}

	err = mySQL.Builder(&st).Table("user").Select("username", "password").One()
	assert.NoError(t, err)
	assert.Equal(t, "test1", st.Username)
	assert.Equal(t, "v123456", st.Password)
}

func TestBuilder_Select(t *testing.T) {
	my := &MySQl{}
	builder := NewBuilder(my, &User{}).Select([]string{"username", "password"})
	assert.Equal(t, 2, len(builder.columns))

	builder.Select("status", "created_at")
	assert.Equal(t, 4, len(builder.columns))
}

func TestNewBuilder(t *testing.T) {
	my := &MySQl{}
	user := &User{}
	assert.Equal(t, NewBuilder(my, user), &Builder{db: my, data: user, from: "user"})
}

func TestBuilder_String(t *testing.T) {
	s := NewBuilder(&MySQl{}, &User{}).Select("username", "password").String()
	assert.NotEqual(t, 0, len(s))
}

func TestBuilder_Table(t *testing.T) {
	builder := NewBuilder(&MySQl{}, &User{}).Table("user")
	assert.Equal(t, "user", builder.from)
}

func TestBuilder_Where(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName, userPathName)
	user := &User{}
	builder := NewBuilder(mySQL, user).Table("user").
		Where("username = ?", "test1").
		Where("status = 1").
		Where(func(builder *Builder) *Builder {
			return builder.Where("status", 1).Where("status", "!=", 10)
		}).
		Where(1).
		Where("status", "between", []int{1, 2}).
		Where("updated_at", "between", "2020-11-22 16:19:11", DateTime()).
		Where("created_at", ">=", "2020-11-14 22:18:37").
		Where("password", "v123456")
	fmt.Printf("%s\n", builder)
	assert.Equal(t, 6, len(builder.wheres))

	err := builder.One()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.UserId)
}

func TestBuilder_Where1(t *testing.T) {
	builder := &Builder{}
	builder.Where(func(builder *Builder) *Builder {
		return builder.Where("status", 1).
			OrWhere("status", 2).
			Where("created_at", "<=", DateTime())
	})

	fmt.Printf("%#v \n", builder.whereFormat())
}

func TestBuilder_Delete(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName, userPathName)
	num, err := mySQL.Builder(&User{}).Where("status", 1).Delete()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), num)
}

func TestBuilder_Update(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName, userPathName)

	var fn BuilderFn = func(builder *Builder) *Builder {
		return builder.Where("created_at", "<=", DateTime()).
			Where("status", "in", 1, 2)
	}

	num, err := mySQL.Builder(&User{Status: 2}).
		Where("status", 1).
		Where(fn).
		Update()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), num)
}

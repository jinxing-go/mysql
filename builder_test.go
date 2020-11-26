package mysql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
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
		Where("status", "between", []int{1, 2}).
		Where("updated_at", "between", "2020-11-22 16:19:11", DateTime()).
		Where("created_at", ">=", "2020-11-14 22:18:37").
		Where("password", "v123456")
	fmt.Printf("%s\n", builder)
	assert.Equal(t, 5, len(builder.wheres))

	err := builder.One()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.UserId)
}

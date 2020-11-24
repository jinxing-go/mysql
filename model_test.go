package mysql

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type User struct {
	UserId    int64  `db:"user_id" json:"user_id"`
	Username  string `db:"username" json:"username"`
	Password  string `db:"password" json:"password"`
	Status    int    `db:"status" json:"status"`
	CreatedAt Time   `db:"created_at" json:"created_at"`
	UpdatedAt Time   `db:"updated_at" json:"updated_at"`
}

func (*User) TableName() string {
	return "user"
}

func (*User) PK() string {
	return "user_id"
}

func (*User) TimestampsValue() interface{} {
	return Time(time.Now())
}

type UserTimestamps struct {
	UserId     int64  `db:"user_id" json:"user_id"`
	Username   string `db:"username" json:"username"`
	CreateTime string `db:"create_time" json:"created_time"`
	UpdateTime string `db:"update_time" json:"update_time"`
}

func (*UserTimestamps) TableName() string {
	return "user_time"
}

func (*UserTimestamps) PK() string {
	return "user_id"
}

func (*UserTimestamps) CreatedAtName() string {
	return "create_time"
}

func (*UserTimestamps) UpdatedAtName() string {
	return "update_time"
}

func (*UserTimestamps) TimestampsValue() interface{} {
	return time.Now().Format(DateLayout)
}

type Log struct {
	UserId     int64  `db:"user_id" json:"user_id"`
	Username   string `db:"username" json:"username"`
	CreateTime string `db:"create_time" json:"created_time"`
	UpdateTime string `db:"update_time" json:"update_time"`
}

func (*Log) TableName() string {
	return "user_time"
}

func (*Log) PK() string {
	return "user_id"
}

func (*Log) AutoTimestamps() bool {
	return false
}

type LoginUser struct {
	UserId    int64     `db:"user_id" json:"user_id"`
	Username  string    `db:"username" json:"username"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (*LoginUser) TableName() string {
	return "login_user"
}

func (*LoginUser) PK() string {
	return "user_id"
}

func TestGetPkValue(t *testing.T) {
	fmt.Printf("%T", GetPKValue(&User{UserId: 1}))
	assert.Equal(t, int64(0), GetPKValue(&User{}))
	assert.Equal(t, int64(1), GetPKValue(&User{UserId: 1}))
}

func TestSetPKValue(t *testing.T) {
	user := &User{UserId: 0}
	SetPKValue(user, 2)
	assert.Equal(t, int64(2), user.UserId)
	SetPKValue(user, 100)
	assert.Equal(t, int64(2), user.UserId)
}

func TestSetCreateAutoTimestamps(t *testing.T) {
	timestamps := &UserTimestamps{}
	auto := SetCreateAutoTimestamps(timestamps)
	assert.Equal(t, true, auto)
	fmt.Printf("%#v \n", timestamps)
	date := Date()
	assert.Equal(t, date, timestamps.CreateTime)
	assert.Equal(t, date, timestamps.UpdateTime)

	log := &Log{}
	auto = SetCreateAutoTimestamps(log)
	fmt.Printf("%#v \n", log)
	assert.Equal(t, false, auto)
}

func TestSetUpdateAutoTimestamps(t *testing.T) {
	timestamps := &UserTimestamps{}
	auto := SetUpdateAutoTimestamps(timestamps)
	assert.Equal(t, true, auto)
	fmt.Printf("%#v \n", timestamps)
	assert.Equal(t, Date(), timestamps.UpdateTime)

	log := &Log{}
	auto = SetUpdateAutoTimestamps(log)
	fmt.Printf("%#v \n", log)
	assert.Equal(t, false, auto)
}

func TestGetCreatedAtColumnName(t *testing.T) {
	user := &User{}
	name := GetCreatedAtColumnName(user)
	assert.Equal(t, CreatedAt, name)

	timestamps := &UserTimestamps{}
	name = GetCreatedAtColumnName(timestamps)
	assert.Equal(t, "create_time", name)
}

func TestGetUpdatedAtColumnName(t *testing.T) {
	user := &User{}
	name := GetUpdatedAtColumnName(user)
	assert.Equal(t, UpdatedAt, name)

	timestamps := &UserTimestamps{}
	name = GetUpdatedAtColumnName(timestamps)
	assert.Equal(t, "update_time", name)
}

func TestGetTimestampsValue(t *testing.T) {
	user := &User{}
	name := GetTimestampsValue(user)
	_, ok := name.(Time)
	assert.Equal(t, true, ok)

	timestamps := &UserTimestamps{}
	name = GetTimestampsValue(timestamps)
	assert.Equal(t, Date(), name.(string))

	loginUser := &LoginUser{}
	timeValue := GetTimestampsValue(loginUser)
	_, isTime := timeValue.(time.Time)
	assert.Equal(t, true, isTime)
}

func TestSetTimestampsValue(t *testing.T) {
	user := &User{}
	value := reflect.ValueOf(user).Elem()
	has := SetStructNameValue(value, "u", int64(1))
	assert.Equal(t, false, has)

	has = SetStructNameValue(value, "user_id", int64(1))
	assert.Equal(t, true, has)
	assert.Equal(t, int64(1), user.UserId)
}

package mysql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type User struct {
	UserId    int64     `db:"user_id" json:"user_id"`
	Username  string    `db:"username" json:"username"`
	Password  string    `db:"password" json:"password"`
	Status    int       `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt Time      `db:"updated_at" json:"updated_at"`
}

func (*User) TableName() string {
	return "user"
}

func (*User) PK() string {
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

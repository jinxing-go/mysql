package mysql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructColumns(t *testing.T) {
	user := StructColumns(&User{UserId: 1}, "db")
	fmt.Printf("%#v\n", user)
	assert.Equal(t, 6, len(user))
}

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

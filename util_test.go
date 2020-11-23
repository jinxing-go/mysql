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

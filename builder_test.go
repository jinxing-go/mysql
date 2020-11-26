package mysql

import (
	"fmt"
	"testing"
)

func TestBuilder_Table(t *testing.T) {
	mySQL := NewTestMySQL(t, examplePathName, userPathName)
	b := NewBuilder(mySQL)
	b.Table("user").
		Select("username", "password").
		Where("username", "like", "username").
		Where("status", "in", 1, 2).
		Where("status", "between", 1, 2).
		Where("username", "jinxing.liu").
		Where("age = ?", 1).Where(fmt.Sprintf(`password = "%s"`, "123")).One(&User{})
	fmt.Printf("%s \n", b)
	fmt.Printf("%#v \n", b.bindings)
}

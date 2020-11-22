package mysql

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	assert.Equal(t, time.Now().Format(DateLayout), Date())
}

func TestDateTime(t *testing.T) {
	assert.Equal(t, time.Now().Format(DateTimeLayout), DateTime())
}

func TestTime_MarshalJSON(t *testing.T) {
	v := Time(time.Now())
	st, err := v.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(`"%s"`, time.Now().Format(DateTimeLayout)), string(st))
}

func TestTime_UnmarshalJSON(t *testing.T) {
	s := fmt.Sprintf(`"%s"`, time.Now().Format(DateTimeLayout))
	var v Time
	err := v.UnmarshalJSON([]byte(s))
	assert.NoError(t, err)
}

func TestTime_String(t *testing.T) {
	v := Time(time.Now())
	assert.Equal(t, time.Now().Format(DateTimeLayout), v.String())
}

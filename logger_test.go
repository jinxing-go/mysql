package mysql

import (
	"testing"
	"time"
)

func TestDefaultLogger_Logger(t *testing.T) {
	logger := &DefaultLogger{}
	logger.Logger(&QueryParams{
		Query: "SELECT * FROM `user` LIMIT 1",
		Start: time.Now(),
		End:   time.Now().Add(1),
	})
}

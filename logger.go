package mysql

import (
	"fmt"
)

type Logger interface {
	Logger(query *QueryParams)
}

type DefaultLogger struct {
}

func (*DefaultLogger) Logger(query *QueryParams) {
	fmt.Println(DateTime())
	fmt.Printf("\t\tQuery: %s\n", query.Query)
	fmt.Printf("\t\tArgs:  %#v\n", query.Args)
	if query.Error != nil {
		fmt.Printf("\t\tError: %#v\n", query.Error)
	}

	fmt.Printf("\t\tTime:  %.4fs\n", query.End.Sub(query.Start).Seconds())
}

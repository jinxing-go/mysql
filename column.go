package mysql

import (
	"reflect"
)

type Column struct {
	Name   string      `json:"name"`
	Value  interface{} `json:"value"`
	IsZero bool        `json:"is_zero"`
}

func StructColumns(data interface{}, tagName string) []*Column {
	v := reflect.ValueOf(data).Elem()
	name := reflect.TypeOf(data).Elem()
	columns := make([]*Column, 0)
	for i, length := 0, v.NumField(); i < length; i++ {
		columns = append(columns, &Column{
			Name:   name.Field(i).Tag.Get(tagName),
			Value:  v.Field(i).Interface(),
			IsZero: v.Field(i).IsZero(),
		})
	}

	return columns
}

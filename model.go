package mysql

import (
	"reflect"
	"time"
)

const (
	CreatedAt = "created_at"
	UpdatedAt = "updated_at"
)

type Model interface {
	TableName() string
	PK() string
}

type CreatedAtName interface {
	CreatedAtName() string
}

type UpdatedAtName interface {
	UpdatedAtName() string
}

type AutoTimestamps interface {
	AutoTimestamps() bool
}

type TimestampsValue interface {
	TimestampsValue() interface{}
}

func GetPKValue(model Model) interface{} {
	return reflect.ValueOf(model).Elem().FieldByName(Studly(model.PK())).Interface()
}

func SetPKValue(model Model, id int64) {
	idField := Studly(model.PK())
	value := reflect.ValueOf(model).Elem().FieldByName(idField)
	if value.IsZero() && value.CanSet() {
		value.SetInt(id)
	}
}

func SetCreateAutoTimestamps(model Model) bool {
	// 如果设置不自动处理时间，那么直接返回false
	if IsAutoTimestamps(model) == false {
		return false
	}

	valueOf := reflect.ValueOf(model).Elem()
	createdAt := GetCreatedAtColumnName(model)
	updatedAt := GetUpdatedAtColumnName(model)
	timeValue := GetTimestampsValue(model)
	SetStructNameValue(valueOf, createdAt, timeValue)
	SetStructNameValue(valueOf, updatedAt, timeValue)
	return true
}

func SetUpdateAutoTimestamps(model Model) bool {
	// 如果设置不自动处理时间，那么直接返回false
	if IsAutoTimestamps(model) == false {
		return false
	}

	updatedAt := GetUpdatedAtColumnName(model)
	timeValue := GetTimestampsValue(model)
	SetStructNameValue(reflect.ValueOf(model).Elem(), updatedAt, timeValue)

	return true
}

// SetStructNameValue 设置结构体指定字段的值
func SetStructNameValue(value reflect.Value, column string, structNameValue interface{}) bool {
	structName := Studly(column)
	if _, ok := value.Type().FieldByName(structName); !ok {
		return false
	}

	valueOf := value.FieldByName(structName)
	if valueOf.IsZero() && valueOf.CanSet() {
		valueOf.Set(reflect.ValueOf(structNameValue))
		return true
	}

	return false
}

// GetTimestampsValue 设置时间值
func GetTimestampsValue(model Model) interface{} {
	// 判断是否实现了自定义处理时间
	if timestampsValue, ok := model.(TimestampsValue); ok {
		return timestampsValue.TimestampsValue()
	}

	return time.Now()
}

// IsAutoTimestamps 是否需要自动处理时间
func IsAutoTimestamps(model Model) bool {
	// 如果设置不自动处理时间，那么直接返回false
	if timestamps, ok := model.(AutoTimestamps); ok && timestamps.AutoTimestamps() == false {
		return false
	}

	return true
}

// GetCreatedAtColumnName 获取创建时间字段名称
func GetCreatedAtColumnName(model Model) string {
	if createdAtColumn, ok := model.(CreatedAtName); ok {
		return createdAtColumn.CreatedAtName()
	}

	return CreatedAt
}

// GetUpdatedAtColumnName 获取修改时间字段名称
func GetUpdatedAtColumnName(model Model) string {
	if updatedAtColumn, ok := model.(UpdatedAtName); ok {
		return updatedAtColumn.UpdatedAtName()
	}

	return UpdatedAt
}

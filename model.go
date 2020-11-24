package mysql

import (
	"reflect"
	"time"
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
	createdValueOf := valueOf.FieldByName(Studly(createdAt))
	if createdValueOf.IsZero() && createdValueOf.CanSet() {
		createdValueOf.Set(reflect.ValueOf(timeValue))
	}

	updatedValueOf := valueOf.FieldByName(Studly(updatedAt))
	if updatedValueOf.IsZero() && updatedValueOf.CanSet() {
		updatedValueOf.Set(reflect.ValueOf(timeValue))
	}

	return true
}

func SetUpdateAutoTimestamps(model Model) bool {
	// 如果设置不自动处理时间，那么直接返回false
	if IsAutoTimestamps(model) == false {
		return false
	}

	updatedAt := GetUpdatedAtColumnName(model)
	timeValue := GetTimestampsValue(model)
	updatedValueOf := reflect.ValueOf(model).Elem().FieldByName(Studly(updatedAt))
	if updatedValueOf.IsZero() && updatedValueOf.CanSet() {
		updatedValueOf.Set(reflect.ValueOf(timeValue))
	}

	return true
}

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

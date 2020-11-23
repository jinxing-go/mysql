package mysql

import "reflect"

type Model interface {
	TableName() string
	PK() string
}

type CreatedAtColumn interface {
	CreatedAtName() string
}

type UpdatedAtColumn interface {
	UpdatedAtName() string
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

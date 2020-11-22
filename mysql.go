package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"reflect"
	"strings"
	"time"
)

const (
	CreatedAt       = "created_at"
	UpdatedAt       = "updated_at"
	CreateEventName = "create"
	UpdateEventName = "update"
	QueryEventName  = "query"
	DeleteEventName = "delete"
)

var (
	CreatedAutoColumns = []string{CreatedAt, UpdatedAt}
	UpdatedAutoColumns = []string{UpdatedAt}
)

type MySQl struct {
	DB      *sqlx.DB
	ShowSql bool
}

type MySQLConfig struct {
	Enable       bool   `toml:"enable" json:"enable"`
	Driver       string `toml:"driver" json:"driver"`
	Dsn          string `toml:"dsn" json:"dsn"`
	MaxOpenConns int    `toml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns int    `toml:"max_idle_conns" json:"max_idle_conns"`
	MaxLifetime  int    `toml:"max_lefttime" json:"max_lefttime"`
	ShowSql      bool   `toml:"show_sql" json:"show_sql"`
}

type QueryParams struct {
	Query string      `json:"query"`
	Args  interface{} `json:"args"`
	Error error       `json:"error"`
	Start time.Time   `json:"start"`
}

type Column struct {
	Name   string      `json:"name"`
	Value  interface{} `json:"value"`
	IsZero bool        `json:"is_zero"`
}

type Model interface {
	TableName() string
	PK() string
}

func NewMySQL(configValue *MySQLConfig) *MySQl {
	db, err := sqlx.Connect(configValue.Driver, configValue.Dsn)
	if err != nil {
		fmt.Println("MySQL connection error: ", err)
		panic(err)
	}

	db.SetMaxIdleConns(configValue.MaxIdleConns)
	db.SetMaxOpenConns(configValue.MaxOpenConns)
	if configValue.MaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(configValue.MaxLifetime) * time.Second)
	}

	return &MySQl{
		DB:      db,
		ShowSql: configValue.ShowSql,
	}
}

// Find 查询数据
func (m *MySQl) Find(model Model, zeroColumn ...string) (err error) {

	where, args := m.toQueryWhere(model, QueryEventName, nil, zeroColumn)
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE %s LIMIT 1", model.TableName(), strings.Join(where, " AND "))

	// 记录日志
	defer func(start time.Time) {
		m.logger(&QueryParams{
			Query: sql,
			Args:  args,
			Error: err,
			Start: start,
		})
	}(time.Now())

	return m.DB.Get(model, sql, args...)
}

// Delete 删除数据
func (m *MySQl) Delete(model Model, zeroColumns ...string) (int64, error) {
	where, args := m.toQueryWhere(model, DeleteEventName, nil, zeroColumns)
	return m.Exec(
		fmt.Sprintf("DELETE FROM `%s` WHERE %s LIMIT 1", model.TableName(), strings.Join(where, " AND ")),
		args...,
	)
}

// Update 修改数据
func (m *MySQl) Update(model Model, zeroColumn ...string) (int64, error) {
	pk := model.PK()
	where, args := m.toQueryWhere(model, UpdateEventName, []string{pk}, zeroColumn)
	args = append(args, m.getPkValue(model, pk))
	return m.Exec(
		fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s` = ?", model.TableName(), strings.Join(where, ", "), pk),
		args...,
	)
}

func (m *MySQl) Create(model Model) (err error) {
	pk := model.PK()
	columns := m.AnalyticStructure(model, CreateEventName)
	fields := make([]string, 0)
	bind := make([]string, 0)
	bindValue := make([]interface{}, 0)
	for _, value := range columns {
		if value.Name != pk && !value.IsZero {
			fields = append(fields, "`"+value.Name+"`")
			bind = append(bind, "?")
			bindValue = append(bindValue, value.Value)
		}
	}

	sql := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		model.TableName(),
		strings.Join(fields, ", "),
		strings.Join(bind, ", "),
	)

	defer func(start time.Time) {
		m.logger(&QueryParams{
			Query: sql,
			Args:  bindValue,
			Error: err,
			Start: start,
		})
	}(time.Now())

	// 执行SQL
	result, err := m.DB.Exec(sql, bindValue...)
	if err != nil {
		return err
	}

	// 获取自增ID
	id, err := result.LastInsertId()
	// 赋值主键值
	m.setPKValue(model, id)

	return err
}

func (m *MySQl) Exec(sql string, args ...interface{}) (i int64, err error) {
	defer func(start time.Time) {
		m.logger(&QueryParams{
			Query: sql,
			Args:  args,
			Error: err,
			Start: start,
		})
	}(time.Now())

	result, err := m.DB.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// toStudly user_id to UserId
func (m *MySQl) toStudly(key string) string {
	// user_id to `user id`
	s := strings.Replace(key, "_", " ", -1)
	// `user id` to `User Id`
	s = strings.Title(s)
	// `User Id` to `UserId`
	return strings.Replace(s, " ", "", -1)
}

func (m *MySQl) getPkValue(data interface{}, key string) interface{} {
	return reflect.ValueOf(data).Elem().FieldByName(m.toStudly(key)).Interface()
}

func (m *MySQl) logger(query *QueryParams) {
	if m.ShowSql {
		fmt.Println(DateTime())
		fmt.Printf("\t\tQuery: %s\n", query.Query)
		fmt.Printf("\t\tArgs:  %#v\n", query.Args)
		if query.Error != nil {
			fmt.Printf("\t\tError: %#v\n", query.Error)
		}

		fmt.Printf("\t\tTime:  %.4fs\n", time.Now().Sub(query.Start).Seconds())
	}
}

func (m *MySQl) AnalyticStructure(data interface{}, eventName string) []*Column {
	v := reflect.ValueOf(data).Elem()
	name := reflect.TypeOf(data).Elem()
	columns := make([]*Column, 0)
	for i, length := 0, v.NumField(); i < length; i++ {
		column := &Column{Name: name.Field(i).Tag.Get("db")}
		isNotHandler := true
		fieldValue := v.Field(i)
		switch eventName {
		case CreateEventName:
			if m.InStringSlice(CreatedAutoColumns, column.Name) {
				m.setTimeValue(column, fieldValue)
				isNotHandler = false
			}
		case UpdateEventName:
			if m.InStringSlice(UpdatedAutoColumns, column.Name) {
				m.setTimeValue(column, fieldValue)
				isNotHandler = false
			}
		}

		if isNotHandler {
			column.Value = v.Field(i).Interface()
			column.IsZero = v.Field(i).IsZero()
		}

		columns = append(columns, column)
	}

	return columns
}

func (m *MySQl) toQueryWhere(data Model, eventName string, exceptColumn, joinColumn []string) ([]string, []interface{}) {
	where := make([]string, 0)
	args := make([]interface{}, 0)

	// 处理需要加入和排除的字段
	except, join := m.toMap(exceptColumn), m.toMap(joinColumn)

	columns := m.AnalyticStructure(data, eventName)
	for _, v := range columns {
		// 先排除
		if _, isExcept := except[v.Name]; isExcept {
			continue
		}

		// 需要加入
		_, isJoin := join[v.Name]
		if !v.IsZero || isJoin {
			where = append(where, fmt.Sprintf("`%s` = ?", v.Name))
			args = append(args, v.Value)
		}
	}

	return where, args
}

func (m *MySQl) toMap(str []string) map[string]bool {
	data := make(map[string]bool)
	for _, key := range str {
		data[key] = true
	}

	return data
}

func (m *MySQl) InStringSlice(strSlice []string, need string) bool {
	if len(strSlice) == 0 {
		return false
	}

	for _, v := range strSlice {
		if v == need {
			return true
		}
	}

	return false
}

func (m *MySQl) setTimeValue(column *Column, value reflect.Value) {
	timeNow := time.Now()
	column.Value = timeNow
	column.IsZero = false
	if value.CanSet() {
		switch value.Interface().(type) {
		case time.Time:
			value.Set(reflect.ValueOf(timeNow))
		case Time:
			value.Set(reflect.ValueOf(Time(timeNow)))
		}
	}
}

func (m *MySQl) setPKValue(model Model, id int64) {
	idField := m.toStudly(model.PK())
	value := reflect.ValueOf(model).Elem().FieldByName(idField)
	if value.IsZero() && value.CanSet() {
		value.SetInt(id)
	}
}

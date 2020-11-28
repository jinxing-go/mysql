package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MySQl struct {
	DB      *sqlx.DB
	ShowSql bool
	Logger  Logger
}

type Config struct {
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
	End   time.Time   `json:"end"`
}

func NewMySQL(configValue *Config) *MySQl {
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
		Logger:  &DefaultLogger{},
	}
}

func (m *MySQl) Get(data interface{}, query string, args ...interface{}) (err error) {

	var (
		queryString string
		bindings    []interface{}
	)

	queryString, bindings, err = sqlx.In(query, args...)
	if err != nil {
		return err
	}

	// 记录日志
	defer func(start time.Time) {
		m.logger(&QueryParams{
			Query: queryString,
			Args:  bindings,
			Error: err,
			Start: start,
			End:   time.Now(),
		})
	}(time.Now())

	return m.DB.Get(data, queryString, bindings...)
}

func (m *MySQl) Select(data interface{}, query string, args ...interface{}) (err error) {
	var (
		queryString string
		bindings    []interface{}
	)

	queryString, bindings, err = sqlx.In(query, args...)
	if err != nil {
		return err
	}

	// 记录日志
	defer func(start time.Time) {
		m.logger(&QueryParams{
			Query: queryString,
			Args:  bindings,
			Error: err,
			Start: start,
			End:   time.Now(),
		})
	}(time.Now())

	return m.DB.Select(data, queryString, bindings...)
}

func (m *MySQl) Builder(data interface{}) *Builder {
	return NewBuilder(m, data)
}

// Find 查询数据
func (m *MySQl) Find(model Model, zeroColumn ...string) (err error) {
	where, args := ToQueryWhere(model, nil, zeroColumn)
	query := fmt.Sprintf("SELECT * FROM `%s` WHERE %s LIMIT 1", model.TableName(), strings.Join(where, " AND "))
	return m.Get(model, query, args...)
}

func (m *MySQl) FindAll(models interface{}, where string, args ...interface{}) error {
	model, err := GetModel(models)
	if err != nil {
		panic(err.Error())
	}

	if where != "" {
		where = "WHERE " + where
	}

	return m.Select(models, fmt.Sprintf("SELECT * FROM `%s` %s", model.TableName(), where), args...)
}

// 创建数据
func (m *MySQl) Create(model Model) (err error) {
	pk := model.PK()
	SetCreateAutoTimestamps(model)
	columns := StructColumns(model, "db")
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

	query := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		model.TableName(),
		strings.Join(fields, ", "),
		strings.Join(bind, ", "),
	)

	defer func(start time.Time) {
		m.logger(&QueryParams{
			Query: query,
			Args:  bindValue,
			Error: err,
			Start: start,
			End:   time.Now(),
		})
	}(time.Now())

	// 执行SQL
	var result sql.Result
	result, err = m.DB.Exec(query, bindValue...)
	if err != nil {
		return err
	}

	// 获取自增ID
	var id int64
	id, err = result.LastInsertId()
	// 赋值主键值
	SetPKValue(model, id)

	return err
}

// Update 修改数据
func (m *MySQl) Update(model Model, zeroColumn ...string) (int64, error) {
	pk := model.PK()
	SetUpdateAutoTimestamps(model)
	where, args := ToQueryWhere(model, []string{pk}, zeroColumn)
	args = append(args, GetPKValue(model))
	return m.Exec(
		fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s` = ? LIMIT 1", model.TableName(), strings.Join(where, ", "), pk),
		args...,
	)
}

// Delete 删除数据
func (m *MySQl) Delete(model Model, zeroColumns ...string) (int64, error) {
	where, args := ToQueryWhere(model, nil, zeroColumns)
	return m.Exec(
		fmt.Sprintf("DELETE FROM `%s` WHERE %s LIMIT 1", model.TableName(), strings.Join(where, " AND ")),
		args...,
	)
}

func (m *MySQl) Exec(query string, args ...interface{}) (i int64, err error) {

	// IN 处理
	queryString, bindings, err1 := sqlx.In(query, args...)
	if err1 != nil {
		return 0, err1
	}

	// 记录日志
	defer func(start time.Time) {
		m.logger(&QueryParams{
			Query: queryString,
			Args:  bindings,
			Error: err,
			Start: start,
			End:   time.Now(),
		})
	}(time.Now())

	var result sql.Result
	result, err = m.DB.Exec(queryString, bindings...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (m *MySQl) logger(params *QueryParams) {
	if m.ShowSql && m.Logger != nil {
		m.Logger.Logger(params)
	}
}

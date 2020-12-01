package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type IDBSqlx interface {
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Preparex(query string) (*sqlx.Stmt, error)
	Rebind(query string) string
	DriverName() string
}

type MySQl struct {
	db      *sqlx.DB
	tx      *sqlx.Tx
	showSql bool
	log     Logger
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
		db:      db,
		showSql: configValue.ShowSql,
		log:     &DefaultLogger{},
	}
}

func (m *MySQl) DB() IDBSqlx {
	if m.tx != nil {
		return m.tx.Unsafe()
	}

	return m.db.Unsafe()
}

// Transaction 事务处理
func (m *MySQl) Transaction(funName func(mysql *MySQl) error) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}

	if err := funName(&MySQl{tx: tx, showSql: m.showSql, log: m.log}); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// 开启事务
func (m *MySQl) Begin() (*MySQl, error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return nil, err
	}

	return &MySQl{tx: tx, showSql: m.showSql, log: m.log}, nil
}

func (m *MySQl) Commit() error {
	return m.tx.Commit()
}

func (m *MySQl) Rollback() error {
	return m.tx.Rollback()
}

// Get 查询一条数据
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

	return m.DB().Get(data, queryString, bindings...)
}

// Select 查询多条数据
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

	return m.DB().Select(data, queryString, bindings...)
}

// Builder 获取查询对象
func (m *MySQl) Builder(data interface{}) *Builder {
	return NewBuilder(m, data)
}

// Find 查询一条数据
func (m *MySQl) Find(model Model, zeroColumn ...string) (err error) {
	where, args := ToQueryWhere(model, nil, zeroColumn)
	query := fmt.Sprintf("SELECT * FROM `%s` WHERE %s LIMIT 1", model.TableName(), strings.Join(where, " AND "))
	return m.Get(model, query, args...)
}

// FindAll 查询多条数据
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

// Create 创建数据
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
	result, err = m.DB().Exec(query, bindValue...)
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
	result, err = m.DB().Exec(queryString, bindings...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (m *MySQl) ShowSql(isShow bool) *MySQl {
	m.showSql = isShow
	return m
}

func (m *MySQl) Logger(log Logger) *MySQl {
	m.log = log
	return m
}

func (m *MySQl) Close() error {
	if m.db != nil {
		return m.db.Close()
	}

	return nil
}

func (m *MySQl) logger(params *QueryParams) {
	if m.showSql && m.log != nil {
		m.log.Logger(params)
	}
}

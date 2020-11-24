package mysql

import (
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
	End   time.Time   `json:"end"`
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
		Logger:  &DefaultLogger{},
	}
}

// Find 查询数据
func (m *MySQl) Find(model Model, zeroColumn ...string) (err error) {

	where, args := m.toQueryWhere(model, nil, zeroColumn)
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE %s LIMIT 1", model.TableName(), strings.Join(where, " AND "))

	// 记录日志
	defer func(start time.Time) {
		m.logger(&QueryParams{
			Query: sql,
			Args:  args,
			Error: err,
			Start: start,
			End:   time.Now(),
		})
	}(time.Now())

	return m.DB.Get(model, sql, args...)
}

// Delete 删除数据
func (m *MySQl) Delete(model Model, zeroColumns ...string) (int64, error) {
	where, args := m.toQueryWhere(model, nil, zeroColumns)
	return m.Exec(
		fmt.Sprintf("DELETE FROM `%s` WHERE %s LIMIT 1", model.TableName(), strings.Join(where, " AND ")),
		args...,
	)
}

// Update 修改数据
func (m *MySQl) Update(model Model, zeroColumn ...string) (int64, error) {
	pk := model.PK()
	SetUpdateAutoTimestamps(model)
	where, args := m.toQueryWhere(model, []string{pk}, zeroColumn)
	args = append(args, GetPKValue(model))
	return m.Exec(
		fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s` = ?", model.TableName(), strings.Join(where, ", "), pk),
		args...,
	)
}

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
			End:   time.Now(),
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
	SetPKValue(model, id)

	return err
}

func (m *MySQl) Exec(sql string, args ...interface{}) (i int64, err error) {
	defer func(start time.Time) {
		m.logger(&QueryParams{
			Query: sql,
			Args:  args,
			Error: err,
			Start: start,
			End:   time.Now(),
		})
	}(time.Now())

	result, err := m.DB.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (m *MySQl) toQueryWhere(data Model, exceptColumn, joinColumn []string) ([]string, []interface{}) {
	where := make([]string, 0)
	args := make([]interface{}, 0)

	// 处理需要加入和排除的字段
	except, join := SliceToMap(exceptColumn), SliceToMap(joinColumn)

	columns := StructColumns(data, "db")
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

func (m *MySQl) logger(params *QueryParams) {
	if m.ShowSql && m.Logger != nil {
		m.Logger.Logger(params)
	}
}

package mysql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

type where struct {
	column   string
	operator string
	value    interface{}
	boolean  string
}

type Builder struct {
	// 使用的db
	db *MySQl

	data interface{}

	// 查询字段
	columns []string

	// 查询表
	from string

	// 查询条件
	wheres []string

	bindings []interface{}

	// 分组
	groups []string

	// hav 条件
	havings []string

	// 分组
	orders []string

	limit int

	offset int
}

func NewBuilder(db *MySQl, model interface{}) *Builder {
	builder := &Builder{db: db, data: model}
	m, err := GetModel(model)
	if err == nil {
		builder.from = m.TableName()
	}

	return builder
}

func (b *Builder) Where(column string, args ...interface{}) *Builder {
	// 自己写的 status = ? and age = ?
	if strings.Index(column, "?") != -1 {
		b.wheres = append(b.wheres, column)
		b.bindings = append(b.bindings, args...)
		return b
	}

	l := len(args)
	switch l {
	case 1: // Where("status", 1)
		b.wheres = append(b.wheres, fmt.Sprintf("`%s` = ?", column))
		b.bindings = append(b.bindings, args[0])
	case 0: // Where("status = 1")
		b.wheres = append(b.wheres, column)
	default: // Where("status", "in", [1, 2, 3]) or Where("status", "between", 1, 2) or Where("age", ">", 1)
		switch args[0].(string) {
		case "in", "IN", "not in", "NOT IN":
			b.wheres = append(b.wheres, fmt.Sprintf("`%s` %s (?)", column, strings.ToUpper(args[0].(string))))
			if l > 2 {
				b.bindings = append(b.bindings, args[1:])
			} else {
				b.bindings = append(b.bindings, args[1])
			}
		case "between", "BETWEEN", "NOT BETWEEN", "not between":
			if l > 2 {
				b.wheres = append(b.wheres, fmt.Sprintf("`%s` %s ? AND ?", column, strings.ToUpper(args[0].(string))))
				b.bindings = append(b.bindings, args[1:]...)
			}
		default:
			b.wheres = append(b.wheres, fmt.Sprintf("`%s` %s ?", column, strings.ToUpper(args[0].(string))))
			b.bindings = append(b.bindings, args[1])
		}
	}

	return b
}

func (b *Builder) Select(column interface{}, columns ...string) *Builder {
	if v, ok := column.([]string); ok {
		b.columns = append(b.columns, v...)
	} else if v, ok := column.(string); ok {
		b.columns = append(b.columns, v)
	}

	if len(columns) > 0 {
		b.columns = append(b.columns, columns...)
	}

	return b
}

func (b *Builder) Table(table string) *Builder {
	b.from = table
	return b
}

func (b *Builder) One() error {
	query, bindings, err := sqlx.In(fmt.Sprintf("%s LIMIT 1", b), b.bindings...)
	if err != nil {
		return err
	}

	return b.db.Get(b.data, query, bindings...)
}

func (b *Builder) All() error {
	query, bindings, err := sqlx.In(b.String(), b.bindings...)
	if err != nil {
		return err
	}

	return b.db.Select(b.data, query, bindings...)
}

func (b *Builder) String() string {
	if b.columns == nil {
		b.columns = []string{"*"}
	}

	for k, v := range b.columns {
		if strings.Index(v, "*") == -1 {
			b.columns[k] = fmt.Sprintf("`%s`", v)
		}
	}

	var where string
	if b.wheres != nil && len(b.wheres) > 0 {
		where = fmt.Sprintf(" WHERE %s", strings.Join(b.wheres, " AND "))
	}

	return fmt.Sprintf("SELECT %s FROM `%s`%s", strings.Join(b.columns, ", "), b.from, where)
}

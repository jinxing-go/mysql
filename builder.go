package mysql

import (
	"fmt"
	"strings"
)

type BuilderFn func(builder *Builder) *Builder

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

func (b *Builder) OrWhere(column interface{}, args ...interface{}) *Builder {
	return b.toWhere("OR", column, args...)
}

func (b *Builder) Where(column interface{}, args ...interface{}) *Builder {
	return b.toWhere("AND", column, args...)
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
	return b.db.Get(b.data, fmt.Sprintf("%s LIMIT 1", b), b.bindings...)
}

func (b *Builder) All() error {
	return b.db.Select(b.data, b.String(), b.bindings...)
}

func (b *Builder) Update(zeroColumn ...string) (int64, error) {
	setColumns, args := ToQueryWhere(b.data, nil, zeroColumn)
	args = append(args, b.bindings...)
	return b.db.Exec(fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", b.from, strings.Join(setColumns, ","), b.whereFormat()), args...)
}

func (b *Builder) Delete() (int64, error) {
	return b.db.Exec(fmt.Sprintf("DELETE FROM `%s` WHERE %s", b.from, b.whereFormat()), b.bindings...)
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
		where = fmt.Sprintf(" WHERE %s", b.whereFormat())
	}

	return fmt.Sprintf("SELECT %s FROM `%s`%s", strings.Join(b.columns, ", "), b.from, where)
}

func (b *Builder) whereFormat() string {
	str := strings.Join(b.wheres, " ")
	str = strings.TrimLeft(str, "AND ")
	str = strings.TrimLeft(str, "OR ")
	return str
}

func (b *Builder) toWhere(boolean string, column interface{}, args ...interface{}) *Builder {

	// 函数执行
	if fn, ok := column.(func(builder *Builder) *Builder); ok {
		builder := fn(&Builder{})
		b.wheres = append(b.wheres, fmt.Sprintf("%s (%s)", boolean, builder.whereFormat()))
		b.bindings = append(b.bindings, builder.bindings...)
		return b
	} else if fn, ok := column.(BuilderFn); ok {
		builder := fn(&Builder{})
		b.wheres = append(b.wheres, fmt.Sprintf("%s (%s)", boolean, builder.whereFormat()))
		b.bindings = append(b.bindings, builder.bindings...)
		return b
	}

	// 字符串处理
	field, ok := column.(string)
	if !ok {
		return b
	}

	// 自己写的 status = ? and age = ?
	if strings.Index(field, "?") != -1 {
		b.wheres = append(b.wheres, fmt.Sprintf("%s (%s)", boolean, field))
		b.bindings = append(b.bindings, args...)
		return b
	}

	l := len(args)
	switch l {
	case 1: // Where("status", 1)
		b.wheres = append(b.wheres, fmt.Sprintf("%s `%s` = ?", boolean, field))
		b.bindings = append(b.bindings, args[0])
	case 0: // Where("status = 1")
		b.wheres = append(b.wheres, fmt.Sprintf("%s (%s)", boolean, field))
	default: // Where("status", "in", [1, 2, 3]) or Where("status", "between", 1, 2) or Where("age", ">", 1)
		switch args[0].(string) {
		case "in", "IN", "not in", "NOT IN":
			b.wheres = append(b.wheres, fmt.Sprintf("%s `%s` %s (?)", boolean, field, strings.ToUpper(args[0].(string))))
			if l > 2 {
				b.bindings = append(b.bindings, args[1:])
			} else {
				b.bindings = append(b.bindings, args[1])
			}
		case "between", "BETWEEN", "NOT BETWEEN", "not between":
			if l > 2 {
				b.wheres = append(b.wheres, fmt.Sprintf("%s `%s` %s ? AND ?", boolean, field, strings.ToUpper(args[0].(string))))
				b.bindings = append(b.bindings, args[1:]...)
			}
		default:
			b.wheres = append(b.wheres, fmt.Sprintf("%s `%s` %s ?", boolean, field, strings.ToUpper(args[0].(string))))
			b.bindings = append(b.bindings, args[1])
		}
	}

	return b
}

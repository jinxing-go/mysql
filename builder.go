package mysql

import (
	"fmt"
	"strconv"
	"strings"
)

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

	joins []string

	// 分组
	groups []string

	// hav 条件
	havings []string

	// 分组
	orders []string

	limit string

	offset string
}

func NewBuilder(db *MySQl, model interface{}) *Builder {
	builder := &Builder{db: db, data: model}
	m, err := GetModel(model)
	if err == nil {
		builder.from = m.TableName()
	}

	return builder
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

func (b *Builder) OrWhere(column interface{}, args ...interface{}) *Builder {
	return b.toWhere("OR", column, args...)
}

func (b *Builder) Where(column interface{}, args ...interface{}) *Builder {
	return b.toWhere("AND", column, args...)
}

func (b *Builder) Join(table, on string, args ...interface{}) *Builder {
	return b.toJoin("JOIN", table, on, args...)
}

func (b *Builder) LeftJoin(table, on string, args ...interface{}) *Builder {
	return b.toJoin("LEFT JOIN", table, on, args...)
}

func (b *Builder) RightJoin(table, on string, args ...interface{}) *Builder {
	return b.toJoin("RIGHT JOIN", table, on, args...)
}

func (b *Builder) OrderBy(column, direction string) *Builder {
	b.orders = append(b.orders, fmt.Sprintf("%s %s", b.warp(column), strings.ToUpper(direction)))
	return b
}

func (b *Builder) GroupBy(groups ...string) *Builder {
	b.groups = append(b.groups, groups...)
	return b
}

func (b *Builder) Having(having string, args ...interface{}) *Builder {
	b.havings = append(b.havings, having)
	b.bindings = append(b.bindings, args...)
	return b
}

func (b *Builder) Limit(limit ...int) *Builder {
	switch len(limit) {
	case 1:
		b.limit = fmt.Sprintf(" LIMIT %d", limit[0])
	case 2:
		b.limit = fmt.Sprintf(" LIMIT %d, %d", limit[0], limit[1])
	}

	return b
}

func (b *Builder) Offset(offset int) *Builder {
	b.offset = fmt.Sprintf(" OFFSET %s", strconv.Itoa(offset))
	return b
}

func (b *Builder) One() error {
	b.limit = " LIMIT 1"
	return b.db.Get(b.data, fmt.Sprintf("%s", b), b.bindings...)
}

func (b *Builder) All() error {
	return b.db.Select(b.data, b.String(), b.bindings...)
}

func (b *Builder) Paginate(page, size int) (int64, error) {
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * size
	b.Offset(offset)
	b.Limit(size)

	// 查询总数
	sql := fmt.Sprintf(
		"SELECT COUNT(*) AS `total` FROM %s%s%s%s%s",
		b.warp(b.from),
		strings.Join(b.joins, ""),
		b.whereFormat(true),
		b.groupByFormat(),
		b.havingFormat(),
	)

	var total int64
	if err := b.db.Get(&total, sql, b.bindings...); err != nil {
		return 0, err
	}

	// 查询数量大于0，才去查询具体数据
	if total > 0 {
		if err := b.All(); err != nil {
			return 0, err
		}
	}

	return total, nil
}

func (b *Builder) Update(zeroColumn ...string) (int64, error) {
	setColumns, args := ToQueryWhere(b.data, nil, zeroColumn)
	args = append(args, b.bindings...)
	return b.db.Exec(fmt.Sprintf("UPDATE %s SET %s%s", b.warp(b.from), strings.Join(setColumns, ","), b.whereFormat(true)), args...)
}

func (b *Builder) Delete() (int64, error) {
	return b.db.Exec(fmt.Sprintf("DELETE FROM %s%s", b.warp(b.from), b.whereFormat(true)), b.bindings...)
}

func (b *Builder) String() string {
	return fmt.Sprintf(
		"SELECT %s FROM %s%s%s%s%s%s%s%s",
		b.columnsFormat(),
		b.warp(b.from),
		strings.Join(b.joins, ""),
		b.whereFormat(true),
		b.groupByFormat(),
		b.havingFormat(),
		b.orderByFormat(),
		b.limit,
		b.offset,
	)
}

func (b *Builder) columnsFormat() string {
	if len(b.columns) == 0 {
		return "*"
	}

	for k, v := range b.columns {
		b.columns[k] = b.warp(v)
	}

	return strings.Join(b.columns, ", ")
}

func (b *Builder) whereFormat(where bool) string {
	if len(b.wheres) == 0 {
		return ""
	}

	str := strings.Join(b.wheres, " ")
	str = strings.TrimLeft(str, "AND ")
	str = strings.TrimLeft(str, "OR ")

	if where {
		return fmt.Sprintf(" WHERE %s", str)
	}

	return str
}
func (b *Builder) groupByFormat() string {
	if b.groups == nil {
		return ""
	}

	for k, v := range b.groups {
		b.groups[k] = b.warp(v)
	}

	return fmt.Sprintf(" GROUP BY %s", strings.Join(b.groups, ", "))
}

func (b *Builder) havingFormat() string {
	if b.havings == nil {
		return ""
	}

	return fmt.Sprintf(" HAVING %s", strings.Join(b.havings, " "))
}

func (b *Builder) orderByFormat() string {
	if b.orders == nil {
		return ""
	}

	return fmt.Sprintf(" ORDER BY %s", strings.Join(b.orders, ", "))
}

func (b *Builder) toJoin(join, table, on string, args ...interface{}) *Builder {
	b.joins = append(b.joins, fmt.Sprintf(" %s %s ON (%s)", join, b.warp(table), on))
	b.bindings = append(b.bindings, args...)
	return b
}

func (b *Builder) toWhere(boolean string, column interface{}, args ...interface{}) *Builder {

	// 函数执行
	if fn, ok := column.(func(builder *Builder) *Builder); ok {
		builder := fn(&Builder{})
		b.wheres = append(b.wheres, fmt.Sprintf("%s (%s)", boolean, builder.whereFormat(false)))
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
		b.wheres = append(b.wheres, fmt.Sprintf("%s %s = ?", boolean, b.warp(field)))
		b.bindings = append(b.bindings, args[0])
	case 0: // Where("status = 1")
		b.wheres = append(b.wheres, fmt.Sprintf("%s (%s)", boolean, field))
	default: // Where("status", "in", [1, 2, 3]) or Where("status", "between", 1, 2) or Where("age", ">", 1)
		switch args[0].(string) {
		case "in", "IN", "not in", "NOT IN":
			b.wheres = append(b.wheres, fmt.Sprintf("%s %s %s (?)", boolean, b.warp(field), strings.ToUpper(args[0].(string))))
			if l > 2 {
				b.bindings = append(b.bindings, args[1:])
			} else {
				b.bindings = append(b.bindings, args[1])
			}
		case "between", "BETWEEN", "NOT BETWEEN", "not between":
			if l > 2 {
				b.wheres = append(b.wheres, fmt.Sprintf("%s %s %s ? AND ?", boolean, b.warp(field), strings.ToUpper(args[0].(string))))
				b.bindings = append(b.bindings, args[1:]...)
			}
		default:
			b.wheres = append(b.wheres, fmt.Sprintf("%s %s %s ?", boolean, b.warp(field), strings.ToUpper(args[0].(string))))
			b.bindings = append(b.bindings, args[1])
		}
	}

	return b
}

func (b *Builder) warp(s string) string {
	// 自己带 `t`.`username`
	if strings.Index(s, "`") != -1 {
		return s
	}

	// table as username
	if strings.Index(s, " as ") != -1 {
		str := strings.Split(s, "as")
		for k, v := range str {
			str[k] = fmt.Sprintf("`%s`", strings.TrimSpace(v))
		}

		return strings.Join(str, " AS ")
	}

	// table AS username
	if strings.Index(s, " AS ") != -1 {
		str := strings.Split(s, "AS")
		for k, v := range str {
			str[k] = fmt.Sprintf("`%s`", strings.TrimSpace(v))
		}

		return strings.Join(str, " AS ")
	}

	// t1 user
	if strings.Index(s, " ") != -1 {
		s = strings.Replace(s, " ", "` `", -1)
	}

	return fmt.Sprintf("`%s`", strings.Replace(s, ".", "`.`", -1))
}

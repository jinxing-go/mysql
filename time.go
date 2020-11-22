package mysql

import "time"

const (
	// DateTimeLayout 默认日期时间格式
	DateTimeLayout = "2006-01-02 15:04:05"

	// DateLayout 默认日期时间格式
	DateLayout = "2006-01-02"

	// TimeZone 时区
	TimeZone = "Asia/Shanghai"
)

type Time time.Time

var loc, _ = time.LoadLocation(TimeZone)

// DateTime 当前日期时间
func DateTime() string {
	return time.Now().In(loc).Format(DateTimeLayout)
}

// Date 当前日期
func Date() string {
	return time.Now().In(loc).Format(DateLayout)
}

func (t *Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(DateTimeLayout)+2)
	b = append(b, '"')
	b = time.Time(*t).In(loc).AppendFormat(b, DateTimeLayout)
	b = append(b, '"')
	return b, nil
}

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+DateTimeLayout+`"`, string(data), loc)
	*t = Time(now)
	return
}

func (t Time) String() string {
	return time.Time(t).In(loc).Format(DateTimeLayout)
}

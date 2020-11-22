package mysql

import "time"

type User struct {
	UserId    int64     `db:"user_id" json:"user_id"`
	Username  string    `db:"username" json:"username"`
	Password  string    `db:"password" json:"password"`
	Status    int       `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt Time      `db:"updated_at" json:"updated_at"`
}

func (*User) TableName() string {
	return "user"
}

func (*User) PK() string {
	return "user_id"
}

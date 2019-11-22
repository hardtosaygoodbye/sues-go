package model

import "time"

type Student struct {
	ID        uint      `gorm:"primary_key"` //ID
	School    string    `gorm:"not null"`    //学校
	Username  string    `gorm:"not null"`    //学号
	Password  string    `gorm:"not null"`    //密码
	CreatedAt time.Time ``                   //创建时间
}

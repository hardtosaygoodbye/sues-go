package main

import "time"

type Course struct {
	Index int `json:"index"`
	Name string `json:"name"`
	Teacher string `json:"teacher"`
	Address string	`json:"address"`
	Week string `json:"week"`
	Time string `json:"time"`
}

type Student struct {
	ID uint `gorm:"primary_key"`
	Num string `gorm:"unique_index"` // 学号
	Password string // 密码
	Name string // 姓名
	Sex string // 性别
	Grade string // 年级
	College string // 学院
	Major string // 专业
	Birthday string // 生日
	IDCard string // 身份证号码
	ComeFrom string // 生源
	Email string // 邮箱
	Phone string // 手机号
	Category string // 学生类别，本科 专科
	Campus string // 校区
	Class string // 班级
	InDate string // 入校时间
	Nation string // 民族
	CreatedAt time.Time // 创建时间
	StudyYear string // 学制
}
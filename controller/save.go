package controller

import (
	"sues-go/driver/mysql"
	"sues-go/model"
)

func saveAccount(t, u, p string) {
	std := model.Student{
		School:   t,
		Username: u,
		Password: p,
	}
	count := 0
	mysql.MainDb.Model(&model.Student{}).Where(&std).Count(&count)
	if count == 0 {
		mysql.MainDb.Create(&std)
	}
}

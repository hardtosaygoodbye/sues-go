package mysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"net/url"
	"strconv"
	"sues-go/config"
	"sues-go/model"
	"time"
)

var DbErr error     // db err instance
var MainDb *gorm.DB // db pool instance

func init() {
	dbConfig := config.GetDbConfig()

	// 生成参数
	v := url.Values{}
	v.Add("charset", dbConfig["CHARSET"]) // 编码
	v.Add("loc", dbConfig["TIMEZONE"])    // 时区
	v.Add("multiStatements", "True")
	v.Add("parseTime", "True")

	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s%s?%s",
		dbConfig["USER"],            // 用户名
		dbConfig["PWD"],             // 密码
		dbConfig["HOST"],            // 主机
		dbConfig["PORT"],            // 端口
		dbConfig["DATABASE_HEAD"],   // 库名头
		dbConfig["DATABASE"],        // 数据库
		dbConfig["DATABASE_FOOTER"], // 库名尾
		v.Encode(),                  // 参数
	)

	// 连接数据库
	MainDb, DbErr = gorm.Open("mysql", dbDSN)
	if DbErr != nil {
		panic("database data source name error: " + DbErr.Error())
	}

	// 连接池最大连接数
	dbMaxOpenConns, _ := strconv.Atoi(dbConfig["MAX_OPEN_CONNS"])
	MainDb.DB().SetMaxOpenConns(dbMaxOpenConns)

	// 连接池最大空闲数
	dbMaxIdleConns, _ := strconv.Atoi(dbConfig["MAX_IDLE_CONNS"])
	MainDb.DB().SetMaxIdleConns(dbMaxIdleConns)

	//连接池链接最长生命周期
	dbMaxLifetimeConns, _ := strconv.Atoi(dbConfig["MAX_LIFETIME_CONNS"])
	MainDb.DB().SetConnMaxLifetime(time.Duration(dbMaxLifetimeConns))

	// Log输出
	MainDb.LogMode(false)

	// 全局禁用表名复数
	MainDb.SingularTable(true)

	// 设置SQL_MODE
	MainDb.Exec("SET GLOBAL sql_mode=\"STRICT_TRANS_TABLES,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION\";")

	// GORM Auto Migrate
	MainDb.AutoMigrate(
		&model.Student{},
	)
}

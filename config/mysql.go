package config

// 数据库配置
func GetDbConfig() map[string]string {
	// 初始化数据库配置map
	dbConfig := make(map[string]string)

	dbConfig["MAX_OPEN_CONNS"] = "20"       // 连接池最大连接数
	dbConfig["MAX_IDLE_CONNS"] = "10"       // 连接池最大空闲数
	dbConfig["MAX_LIFETIME_CONNS"] = "7200" // 连接池链接最长生命周期
	dbConfig["TIMEZONE"] = "Asia/Shanghai"  // 时区

	dbConfig["HOST"] = "127.0.0.1"   // 主机
	dbConfig["PORT"] = "3306"        // 端口
	dbConfig["DATABASE"] = "main"    // 数据库名
	dbConfig["DATABASE_HEAD"] = ""   // 库名头
	dbConfig["DATABASE_FOOTER"] = "" // 库名尾
	dbConfig["USER"] = "root"        // 用户名
	dbConfig["PWD"] = "ch123456"     // 密码
	dbConfig["CHARSET"] = "utf8"     // 编码

	return dbConfig
}

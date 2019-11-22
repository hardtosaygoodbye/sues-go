package main

import (
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"log"
	"sues-go/driver/mysql"
	"sues-go/routes"
	"syscall"
	"time"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	// 路由
	router := routes.SetupRouter()

	// MySQL
	defer mysql.MainDb.Close()

	// 启动服务
	endless.DefaultReadTimeOut = 120 * time.Second
	endless.DefaultWriteTimeOut = 120 * time.Second
	endless.DefaultMaxHeaderBytes = 1 << 20
	server := endless.NewServer(":3000", router)
	server.BeforeBegin = func(add string) {
		log.Printf("Actual Server PID is %d", syscall.Getpid())
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Server err: %v", err)
	}
}

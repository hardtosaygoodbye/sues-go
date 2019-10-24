package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"sues-go/controller"
	"sues-go/controller/middleware"
)

func SetupRouter() *gin.Engine {
	// 初始化路由
	router := gin.New()
	// 修正来自Java的无法识别的Multipart
	router.Use(middleware.FixMultipart())
	// 使用 Logger 中间件
	router.Use(middleware.Logger())
	// 使用 Recovery 中间件
	router.Use(gin.Recovery())
	// 跨域
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	router.Use(cors.New(corsConfig))

	//设置默认路由当访问一个错误网站时返回
	router.NoRoute(notFound)

	routerSUES := router.Group("/sues")
	{
		routerSUES.GET("/courses", controller.GetSUESCourses)
	}

	return router
}

//设置默认路由当访问一个错误网站时返回
func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"detail": "Not Found"})
}

package prom

import "github.com/gin-gonic/gin"

/**
注册该模块下所有的api
*/
func Registry(api *gin.RouterGroup) {
	api.POST("/push", handleLokiPush)
	api.POST("/app", handleAppRegister)
	api.DELETE("/app", handleAppOffline)
	api.GET("/app", handleListApp)
}

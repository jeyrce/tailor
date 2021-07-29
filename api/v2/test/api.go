package test

import "github.com/gin-gonic/gin"

// 预留v2的api注册
func Registry(api *gin.RouterGroup) {
	api.GET("", handleTest)
}

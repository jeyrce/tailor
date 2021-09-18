package api

import (
	"github.com/gin-gonic/gin"
	"woqutech.com/tailor/api/v1/index"
	"woqutech.com/tailor/api/v1/prom"
	"woqutech.com/tailor/api/v2/test"
)

// v1版本api的各个模块注册
func V1(api *gin.RouterGroup) {
	prom.Registry(api.Group("/prom"))
	index.Registry(api.Group("/index"))
}

// v2版本的api各个模块注册
func V2(api *gin.RouterGroup) {
	test.Registry(api.Group("/test"))
}

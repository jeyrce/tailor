package middle

import (
	"github.com/gin-gonic/gin"
	"woqutech.com/tailor/pkg/log"
)

// 注册自定义的中间件
func Registry(e *gin.Engine) *gin.Engine {
	e.Use(
		recovery(log.Logger),   // 自定义的recover
		requestLog(log.Logger), // 自定义的请求记录
	)
	return e
}

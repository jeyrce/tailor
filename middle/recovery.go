package middle

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "woqutech.com/tailor/api/v1"
)

// gin.Recovery() 将会输出一大堆日志, 此处我们自定义简单处理
func recovery(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("服务出现错误,尝试恢复: %v", err)
				v1.ApiFailed(c, new(v1.SystemError))
			}
		}()
		c.Next()
	}
}

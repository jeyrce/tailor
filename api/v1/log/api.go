package log

import "github.com/gin-gonic/gin"

func Registry(api *gin.RouterGroup) {
	api.GET("/summary", handleLogDownload)
}

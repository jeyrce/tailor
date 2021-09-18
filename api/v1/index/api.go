package index

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Registry(api *gin.RouterGroup) {
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api.GET("/ready", handleServerReady)
	api.GET("/version", handleVersion)
}

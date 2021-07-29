package log

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleLogDownload(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"x": "y"})
}

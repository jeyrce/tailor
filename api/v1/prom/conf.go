package prom

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleAddTarget(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"name": "Jeyrce.Lu"})
}

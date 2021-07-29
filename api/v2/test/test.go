package test

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleTest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Author": "Jeyrce.Lu"})
}

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// InitRouter make router
func InitRouter(SECRET_KEY, CurrentPath string, FormMemory int64) *gin.Engine {
	router := gin.Default()
	router.MaxMultipartMemory = FormMemory << 20
	{
		router.GET("/", func(c *gin.Context) {
			method := c.Request.Method
			c.JSON(http.StatusOK, gin.H{"method": method})
		})
	}

	return router
}

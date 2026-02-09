package middleware

import "github.com/gin-gonic/gin"

// SetConfigMiddleWare set config
func SetConfigMiddleWare(SECRET_KEY, CurrentPath, Users_SECRET_KEY string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("secret_key", SECRET_KEY)
		c.Set("users_secret_key", Users_SECRET_KEY)
		c.Set("current_path", CurrentPath)
		c.Writer.Status()
	}
}

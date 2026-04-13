package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/velopert/gin-rest-api-sample/lib/common"
)

func Authorized(c *gin.Context) {
	_, exists := c.Get("user")
	if !exists {
		tokenError, exists := c.Get("token_error")
		if exists {
			switch tokenError {
			case "expired":
				c.JSON(http.StatusUnauthorized, common.JSON{
					"error":   "token_expired",
					"message": "Token has expired",
				})
				c.Abort()
				return
			case "invalid":
				c.JSON(http.StatusUnauthorized, common.JSON{
					"error":   "invalid_token",
					"message": "Invalid token",
				})
				c.Abort()
				return
			}
		}
		c.JSON(http.StatusUnauthorized, common.JSON{
			"error":   "token_missing",
			"message": "Token is missing",
		})
		c.Abort()
		return
	}
}

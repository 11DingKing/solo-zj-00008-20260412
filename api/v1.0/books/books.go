package books

import (
	"github.com/gin-gonic/gin"
	"github.com/velopert/gin-rest-api-sample/lib/middlewares"
)

func ApplyRoutes(r *gin.RouterGroup) {
	books := r.Group("/books")
	{
		books.POST("/", middlewares.Authorized, create)
		books.GET("/", list)
		books.GET("/:id", read)
		books.DELETE("/:id", middlewares.Authorized, remove)
		books.PATCH("/:id", middlewares.Authorized, update)
	}
}

package books

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/velopert/gin-rest-api-sample/database/models"
	"github.com/velopert/gin-rest-api-sample/lib/common"
)

type Book = models.Book
type User = models.User
type JSON = common.JSON

func create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	type RequestBody struct {
		Title  string `json:"title" binding:"required"`
		Author string `json:"author" binding:"required"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, JSON{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
		return
	}

	user := c.MustGet("user").(User)

	book := Book{
		Title:  body.Title,
		Author: body.Author,
		User:   user,
	}

	db.NewRecord(book)
	db.Create(&book)

	db.Preload("User").First(&book, book.ID)

	c.JSON(http.StatusOK, book.Serialize())
}

func list(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	cursor := c.Query("cursor")
	recent := c.Query("recent")

	var books []Book

	if cursor == "" {
		if err := db.Preload("User").Limit(10).Order("id desc").Find(&books).Error; err != nil {
			c.JSON(http.StatusInternalServerError, JSON{
				"error":   "database_error",
				"message": "Failed to fetch books",
			})
			return
		}
	} else {
		condition := "id < ?"
		if recent == "1" {
			condition = "id > ?"
		}
		if err := db.Preload("User").Limit(10).Order("id desc").Where(condition, cursor).Find(&books).Error; err != nil {
			c.JSON(http.StatusInternalServerError, JSON{
				"error":   "database_error",
				"message": "Failed to fetch books",
			})
			return
		}
	}

	length := len(books)
	serialized := make([]JSON, length, length)

	for i := 0; i < length; i++ {
		serialized[i] = books[i].Serialize()
	}

	c.JSON(http.StatusOK, serialized)
}

func read(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var book Book
	if err := db.Preload("User").Where("id = ?", id).First(&book).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, JSON{
				"error":   "not_found",
				"message": "Book not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, JSON{
			"error":   "database_error",
			"message": "Failed to fetch book",
		})
		return
	}

	c.JSON(http.StatusOK, book.Serialize())
}

func remove(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	user := c.MustGet("user").(User)

	var book Book
	if err := db.Where("id = ?", id).First(&book).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, JSON{
				"error":   "not_found",
				"message": "Book not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, JSON{
			"error":   "database_error",
			"message": "Failed to fetch book",
		})
		return
	}

	if book.UserID != user.ID {
		c.JSON(http.StatusForbidden, JSON{
			"error":   "forbidden",
			"message": "You are not allowed to delete this book",
		})
		return
	}

	db.Delete(&book)
	c.Status(http.StatusNoContent)
}

func update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	user := c.MustGet("user").(User)

	type RequestBody struct {
		Title  string `json:"title"`
		Author string `json:"author"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, JSON{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
		return
	}

	var book Book
	if err := db.Preload("User").Where("id = ?", id).First(&book).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, JSON{
				"error":   "not_found",
				"message": "Book not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, JSON{
			"error":   "database_error",
			"message": "Failed to fetch book",
		})
		return
	}

	if book.UserID != user.ID {
		c.JSON(http.StatusForbidden, JSON{
			"error":   "forbidden",
			"message": "You are not allowed to update this book",
		})
		return
	}

	if body.Title != "" {
		book.Title = body.Title
	}
	if body.Author != "" {
		book.Author = body.Author
	}

	db.Save(&book)
	c.JSON(http.StatusOK, book.Serialize())
}

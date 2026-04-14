package models

import (
	"github.com/jinzhu/gorm"
	"github.com/velopert/gin-rest-api-sample/lib/common"
)

type Book struct {
	gorm.Model
	Title  string
	Author string
	User   User `gorm:"foreignkey:UserID"`
	UserID uint
}

func (b Book) Serialize() common.JSON {
	return common.JSON{
		"id":         b.ID,
		"title":      b.Title,
		"author":     b.Author,
		"user":       b.User.Serialize(),
		"created_at": b.CreatedAt,
	}
}

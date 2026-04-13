package auth

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/velopert/gin-rest-api-sample/database/models"
	"github.com/velopert/gin-rest-api-sample/lib/common"
	"github.com/velopert/gin-rest-api-sample/lib/middlewares"
	"golang.org/x/crypto/bcrypt"
)

type User = models.User

func hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func checkHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func register(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	type RequestBody struct {
		Username    string `json:"username" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
		Password    string `json:"password" binding:"required"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(400)
		return
	}

	var exists User
	if err := db.Where("username = ?", body.Username).First(&exists).Error; err == nil {
		c.AbortWithStatus(409)
		return
	}

	hash, hashErr := hash(body.Password)
	if hashErr != nil {
		c.AbortWithStatus(500)
		return
	}

	user := User{
		Username:     body.Username,
		DisplayName:  body.DisplayName,
		PasswordHash: hash,
	}

	db.NewRecord(user)
	db.Create(&user)

	serialized := user.Serialize()
	token, _ := middlewares.GenerateToken(serialized)
	c.SetCookie("token", token, 60*60*24, "/", "", false, true)

	c.JSON(200, common.JSON{
		"user":  user.Serialize(),
		"token": token,
	})
}

func login(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	type RequestBody struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(400)
		return
	}

	var user User
	if err := db.Where("username = ?", body.Username).First(&user).Error; err != nil {
		c.AbortWithStatus(404)
		return
	}

	if !checkHash(body.Password, user.PasswordHash) {
		c.AbortWithStatus(401)
		return
	}

	serialized := user.Serialize()
	token, _ := middlewares.GenerateToken(serialized)

	c.SetCookie("token", token, 60*60*24, "/", "", false, true)

	c.JSON(200, common.JSON{
		"user":  user.Serialize(),
		"token": token,
	})
}

func check(c *gin.Context) {
	userRaw, ok := c.Get("user")
	if !ok {
		c.AbortWithStatus(401)
		return
	}

	user := userRaw.(User)

	tokenExpireRaw, ok := c.Get("token_expire")
	if !ok {
		c.AbortWithStatus(401)
		return
	}

	tokenExpire := tokenExpireRaw.(int64)
	now := time.Now().Unix()
	diff := tokenExpire - now

	fmt.Println(diff)
	if diff < 60*60*12 {
		token, _ := middlewares.GenerateToken(user.Serialize())
		c.SetCookie("token", token, 60*60*24, "/", "", false, true)
		c.JSON(200, common.JSON{
			"token": token,
			"user":  user.Serialize(),
		})
		return
	}

	c.JSON(200, common.JSON{
		"token": nil,
		"user":  user.Serialize(),
	})
}

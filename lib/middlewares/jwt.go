package middlewares

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/velopert/gin-rest-api-sample/database/models"
	"github.com/velopert/gin-rest-api-sample/lib/common"
)

var secretKey []byte

func init() {
	pwd, _ := os.Getwd()
	keyPath := pwd + "/jwtsecret.key"

	key, readErr := ioutil.ReadFile(keyPath)
	if readErr != nil {
		panic("failed to load secret key file")
	}
	secretKey = key
}

type TokenError struct {
	Type    string
	Message string
}

func (e *TokenError) Error() string {
	return e.Message
}

func validateToken(tokenString string) (common.JSON, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return common.JSON{}, &TokenError{Type: "expired", Message: "token has expired"}
		}
		return common.JSON{}, &TokenError{Type: "invalid", Message: "invalid token"}
	}

	if !token.Valid {
		return common.JSON{}, &TokenError{Type: "invalid", Message: "invalid token"}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return common.JSON{}, &TokenError{Type: "invalid", Message: "invalid token claims"}
	}

	return common.JSON(claims), nil
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			authorization := c.Request.Header.Get("Authorization")
			if authorization == "" {
				c.Next()
				return
			}
			sp := strings.Split(authorization, "Bearer ")
			if len(sp) < 2 {
				c.Next()
				return
			}
			tokenString = sp[1]
		}

		tokenData, err := validateToken(tokenString)
		if err != nil {
			var tokenErr *TokenError
			if errors.As(err, &tokenErr) {
				c.Set("token_error", tokenErr.Type)
			}
			c.Next()
			return
		}

		var user models.User
		userData, ok := tokenData["user"].(map[string]interface{})
		if !ok {
			c.Set("token_error", "invalid")
			c.Next()
			return
		}
		user.Read(common.JSON(userData))

		c.Set("user", user)
		exp, ok := tokenData["exp"].(float64)
		if ok {
			c.Set("token_expire", int64(exp))
		}
		c.Next()
	}
}

func generateToken(data common.JSON) (string, error) {
	date := time.Now().Add(time.Hour * 24)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": data,
		"exp":  date.Unix(),
	})

	pwd, _ := os.Getwd()
	keyPath := pwd + "/jwtsecret.key"

	key, readErr := ioutil.ReadFile(keyPath)
	if readErr != nil {
		return "", readErr
	}
	tokenString, err := token.SignedString(key)
	return tokenString, err
}

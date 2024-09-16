package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetPaginationParams(c *gin.Context) (limit int, offset int) {
	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 || limit > 50 {
		limit = 5
	}

	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}

var jwtSecret = []byte("secret_key")

// ValidateToken проверяет валидность JWT токена.
func ValidateToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return false, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp := int64(claims["exp"].(float64))
		if exp < time.Now().Unix() {
			return false, errors.New("token expired")
		}
		return true, nil
	}

	return false, errors.New("invalid token")
}

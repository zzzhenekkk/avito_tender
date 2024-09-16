// internal/middlewares/auth_middleware.go
package middlewares

import (
	"net/http"
	"strings"
	"tender_management_api/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"reason": "Отсутствует токен авторизации"})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"reason": "Неверный формат токена авторизации"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Проверка валидности токена
		isValid, err := utils.ValidateToken(token)
		if err != nil || !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{"reason": "Недействительный токен"})
			c.Abort()
			return
		}

		c.Next()
	}
}

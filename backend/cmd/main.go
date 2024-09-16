package main

import (
	"log"
	"net/http"
	"tender_management_api/internal/config"
	"tender_management_api/internal/database"
	"tender_management_api/internal/middlewares"
	"tender_management_api/internal/routers"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Не удалось загрузить конфигурацию: ", err)
	}

	database.ConnectDatabase(cfg)

	router := gin.Default()

	// Маршрут для проверки доступности сервера
	router.GET("/api/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Группа маршрутов с префиксом /api и middleware для аутентификации
	api := router.Group("/api")
	api.Use(middlewares.AuthMiddleware())

	// Инициализация маршрутов
	routers.InitTenderRoutes(api)
	routers.InitBidRoutes(api)

	// Запуск сервера
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatal("Не удалось запустить сервер: ", err)
	}
}

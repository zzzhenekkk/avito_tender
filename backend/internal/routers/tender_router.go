package routers

import (
	"tender_management_api/internal/controllers"

	"github.com/gin-gonic/gin"
)

func InitTenderRoutes(router *gin.RouterGroup) {
	router.GET("/tenders", controllers.GetTenders)
	router.POST("/tenders/new", controllers.CreateTender)
	router.GET("/tenders/my", controllers.GetUserTenders)
	router.GET("/tenders/:tenderId/status", controllers.GetTenderStatus)
	router.PUT("/tenders/:tenderId/status", controllers.UpdateTenderStatus)
	router.PATCH("/tenders/:tenderId/edit", controllers.EditTender)
	router.PUT("/tenders/:tenderId/rollback/:version", controllers.RollbackTender)
}

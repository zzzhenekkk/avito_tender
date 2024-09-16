package routers

import (
	"tender_management_api/internal/controllers"

	"github.com/gin-gonic/gin"
)

func InitBidRoutes(router *gin.RouterGroup) {
	router.POST("/bids/new", controllers.CreateBid)
	router.GET("/bids/my", controllers.GetUserBids)
	router.GET("/bids/:tenderId/list", controllers.GetBidsForTender)
	router.GET("/bids/:bidId/status", controllers.GetBidStatus)
	router.PUT("/bids/:bidId/status", controllers.UpdateBidStatus)
	router.PATCH("/bids/:bidId/edit", controllers.EditBid)
	router.PUT("/bids/:bidId/rollback/:version", controllers.RollbackBid)
	router.PUT("/bids/:bidId/submit_decision", controllers.SubmitBidDecision)
	router.PUT("/bids/:bidId/feedback", controllers.SubmitBidFeedback)
	router.GET("/bids/:tenderId/reviews", controllers.GetBidReviews)
}

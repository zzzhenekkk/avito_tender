package controllers

import (
	"net/http"
	"strconv"
	"tender_management_api/internal/database"
	"tender_management_api/internal/models"
	"tender_management_api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateBidInput struct {
	Name            string           `json:"name" binding:"required,max=100"`
	Description     string           `json:"description" binding:"required,max=500"`
	Status          models.BidStatus `json:"status" binding:"required,oneof=Created Published Canceled Approved Rejected"`
	TenderID        uuid.UUID        `json:"tenderId" binding:"required"`
	OrganizationID  uuid.UUID        `json:"organizationId" binding:"required"`
	CreatorUsername string           `json:"creatorUsername" binding:"required"`
}

func CreateBid(c *gin.Context) {
	var input CreateBidInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := database.DB.Where("username = ?", input.CreatorUsername).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Пользователь не существует или некорректен"})
		return
	}

	// Проверка существования тендера
	var tender models.Tender
	if err := database.DB.Where("id = ?", input.TenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Тендер не найден"})
		return
	}

	// Проверка, является ли пользователь ответственным за организацию
	var orgResp models.OrganizationResponsible
	if err := database.DB.Where("organization_id = ? AND user_id = ?", input.OrganizationID, user.ID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Недостаточно прав для выполнения действия"})
		return
	}

	bid := models.Bid{
		Name:        input.Name,
		Description: input.Description,
		Status:      input.Status,
		TenderID:    input.TenderID,
		AuthorType:  models.BidAuthorTypeOrganization,
		AuthorID:    input.OrganizationID,
		Version:     1,
	}

	if err := database.DB.Create(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	// Сохраняем версию предложения
	bidVersion := models.BidVersion{
		BidID:       bid.ID,
		Version:     bid.Version,
		Name:        bid.Name,
		Description: bid.Description,
	}
	if err := database.DB.Create(&bidVersion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bid)
}

func GetUserBids(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Username is required"})
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Пользователь не существует или некорректен"})
		return
	}

	var bids []models.Bid
	var total int64

	limit, offset := utils.GetPaginationParams(c)

	database.DB.Model(&models.Bid{}).
		Where("author_id = ?", user.ID).
		Count(&total).
		Limit(limit).Offset(offset).Order("name ASC").Find(&bids)

	c.JSON(http.StatusOK, bids)
}

func GetBidStatus(c *gin.Context) {
	bidID := c.Param("bidId")
	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Параметр 'username' обязателен"})
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Пользователь не существует или некорректен"})
		return
	}

	// Проверка существования предложения
	var bid models.Bid
	if err := database.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Предложение не найдено"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": bid.Status})
}

func UpdateBidStatus(c *gin.Context) {
	bidID := c.Param("bidId")
	username := c.Query("username")
	status := c.Query("status")

	if username == "" || status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Параметры 'username' и 'status' обязательны"})
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Пользователь не существует или некорректен"})
		return
	}

	// Проверка существования предложения
	var bid models.Bid
	if err := database.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Предложение не найдено"})
		return
	}

	// Проверка прав доступа (например, только автор может изменить статус)
	if bid.AuthorID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Недостаточно прав для выполнения действия"})
		return
	}

	// Проверка корректности статуса
	validStatuses := []models.BidStatus{
		models.BidStatusCreated,
		models.BidStatusPublished,
		models.BidStatusCanceled,
		models.BidStatusApproved,
		models.BidStatusRejected,
	}
	isValidStatus := false
	for _, s := range validStatuses {
		if s == models.BidStatus(status) {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Некорректный статус предложения"})
		return
	}

	// Обновление статуса предложения
	bid.Status = models.BidStatus(status)
	if err := database.DB.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bid)
}

func EditBid(c *gin.Context) {
	bidID := c.Param("bidId")
	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Параметр 'username' обязателен"})
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Пользователь не существует или некорректен"})
		return
	}

	// Проверка существования предложения
	var bid models.Bid
	if err := database.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Предложение не найдено"})
		return
	}

	// Проверка прав доступа (только автор может редактировать)
	if bid.AuthorID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Недостаточно прав для выполнения действия"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}

	// Обновление предложения
	input["version"] = bid.Version + 1

	if err := database.DB.Model(&bid).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	// Сохранение версии предложения
	bidVersion := models.BidVersion{
		BidID:       bid.ID,
		Version:     bid.Version,
		Name:        bid.Name,
		Description: bid.Description,
	}
	if err := database.DB.Create(&bidVersion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bid)
}

func RollbackBid(c *gin.Context) {
	bidID := c.Param("bidId")
	versionParam := c.Param("version")
	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Параметр 'username' обязателен"})
		return
	}

	version, err := strconv.Atoi(versionParam)
	if err != nil || version < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Некорректный номер версии"})
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Пользователь не существует или некорректен"})
		return
	}

	// Проверка существования предложения
	var bid models.Bid
	if err := database.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Предложение не найдено"})
		return
	}

	// Проверка прав доступа
	if bid.AuthorID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Недостаточно прав для выполнения действия"})
		return
	}

	// Получение версии предложения
	var bidVersion models.BidVersion
	if err := database.DB.Where("bid_id = ? AND version = ?", bid.ID, version).First(&bidVersion).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Версия не найдена"})
		return
	}

	// Откат предложения
	bid.Name = bidVersion.Name
	bid.Description = bidVersion.Description
	bid.Version += 1

	if err := database.DB.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	// Сохранение новой версии
	newBidVersion := models.BidVersion{
		BidID:       bid.ID,
		Version:     bid.Version,
		Name:        bid.Name,
		Description: bid.Description,
	}
	if err := database.DB.Create(&newBidVersion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bid)
}

func SubmitBidDecision(c *gin.Context) {
	bidID := c.Param("bidId")
	username := c.Query("username")
	decision := c.Query("decision")

	if username == "" || decision == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Параметры 'username' и 'decision' обязательны"})
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Пользователь не существует или некорректен"})
		return
	}

	// Проверка существования предложения
	var bid models.Bid
	if err := database.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Предложение не найдено"})
		return
	}

	// Проверка прав доступа (ответственный за тендер)
	var tender models.Tender
	if err := database.DB.Where("id = ?", bid.TenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Тендер не найден"})
		return
	}

	var orgResp models.OrganizationResponsible
	if err := database.DB.Where("organization_id = ? AND user_id = ?", tender.OrganizationID, user.ID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Недостаточно прав для выполнения действия"})
		return
	}

	// Проверка корректности решения
	if decision != string(models.BidDecisionApproved) && decision != string(models.BidDecisionRejected) {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Некорректное решение"})
		return
	}

	// Логика согласования предложения
	// Нужно реализовать процесс кворума согласно бизнес-логике

	// Для упрощения примера:
	if decision == string(models.BidDecisionApproved) {
		bid.Status = models.BidStatusApproved
		// При согласовании предложения тендер автоматически закрывается
		tender.Status = models.TenderStatusClosed
		if err := database.DB.Save(&tender).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
			return
		}
	} else {
		bid.Status = models.BidStatusRejected
	}

	if err := database.DB.Save(&bid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bid)
}

func SubmitBidFeedback(c *gin.Context) {
	bidID := c.Param("bidId")
	username := c.Query("username")
	bidFeedback := c.Query("bidFeedback")

	if username == "" || bidFeedback == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Параметры 'username' и 'bidFeedback' обязательны"})
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Пользователь не существует или некорректен"})
		return
	}

	// Проверка существования предложения
	var bid models.Bid
	if err := database.DB.Where("id = ?", bidID).First(&bid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Предложение не найдено"})
		return
	}

	// Проверка прав доступа (ответственный за тендер)
	var tender models.Tender
	if err := database.DB.Where("id = ?", bid.TenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Тендер не найден"})
		return
	}

	var orgResp models.OrganizationResponsible
	if err := database.DB.Where("organization_id = ? AND user_id = ?", tender.OrganizationID, user.ID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Недостаточно прав для выполнения действия"})
		return
	}

	// Создание отзыва
	feedback := models.BidFeedback{
		BidID:    bid.ID,
		Feedback: bidFeedback,
	}

	if err := database.DB.Create(&feedback).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bid)
}

func GetBidReviews(c *gin.Context) {
	tenderID := c.Param("tenderId")
	authorUsername := c.Query("authorUsername")
	requesterUsername := c.Query("requesterUsername")

	if authorUsername == "" || requesterUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Параметры 'authorUsername' и 'requesterUsername' обязательны"})
		return
	}

	// Проверка существования пользователя-запросчика
	var requester models.User
	if err := database.DB.Where("username = ?", requesterUsername).First(&requester).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Пользователь-запросчик не существует или некорректен"})
		return
	}

	// Проверка существования автора
	var author models.User
	if err := database.DB.Where("username = ?", authorUsername).First(&author).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Автор не существует или некорректен"})
		return
	}

	// Проверка существования тендера
	var tender models.Tender
	if err := database.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Тендер не найден"})
		return
	}

	// Проверка прав доступа (запросчик должен быть ответственным за тендер)
	var orgResp models.OrganizationResponsible
	if err := database.DB.Where("organization_id = ? AND user_id = ?", tender.OrganizationID, requester.ID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Недостаточно прав для выполнения действия"})
		return
	}

	// Получение отзывов на предложения автора
	limit, offset := utils.GetPaginationParams(c)

	var reviews []models.BidFeedback
	database.DB.Joins("JOIN bids ON bid_feedbacks.bid_id = bids.id").
		Where("bids.author_id = ?", author.ID).
		Limit(limit).Offset(offset).
		Find(&reviews)

	c.JSON(http.StatusOK, reviews)
}

func GetBidsForTender(c *gin.Context) {
	tenderID := c.Param("tenderId")
	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Параметр 'username' обязателен"})
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": "Пользователь не существует или некорректен"})
		return
	}

	// Проверка существования тендера
	var tender models.Tender
	if err := database.DB.Where("id = ?", tenderID).First(&tender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Тендер не найден"})
		return
	}

	// Проверка прав доступа
	var orgResp models.OrganizationResponsible
	if err := database.DB.Where("organization_id = ? AND user_id = ?", tender.OrganizationID, user.ID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Недостаточно прав для выполнения действия"})
		return
	}

	limit, offset := utils.GetPaginationParams(c)

	// Получение списка предложений для указанного тендера
	var bids []models.Bid
	database.DB.Where("tender_id = ?", tenderID).
		Limit(limit).Offset(offset).
		Order("name ASC").Find(&bids)

	c.JSON(http.StatusOK, bids)
}

// internal/controllers/tender_controller.go
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

type CreateTenderInput struct {
	Name            string              `json:"name" binding:"required,max=100"`
	Description     string              `json:"description" binding:"required,max=500"`
	ServiceType     models.ServiceType  `json:"serviceType" binding:"required,oneof=Construction Delivery Manufacture"`
	Status          models.TenderStatus `json:"status" binding:"required,oneof=Created Published Closed"`
	OrganizationID  uuid.UUID           `json:"organizationId" binding:"required"`
	CreatorUsername string              `json:"creatorUsername" binding:"required"`
}

func CreateTender(c *gin.Context) {
	var input CreateTenderInput
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

	// Проверка, является ли пользователь ответственным за организацию
	var orgResp models.OrganizationResponsible
	if err := database.DB.Where("organization_id = ? AND user_id = ?", input.OrganizationID, user.ID).First(&orgResp).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"reason": "Недостаточно прав для выполнения действия"})
		return
	}

	tender := models.Tender{
		Name:           input.Name,
		Description:    input.Description,
		ServiceType:    input.ServiceType,
		Status:         input.Status,
		OrganizationID: input.OrganizationID,
		Version:        1,
	}

	if err := database.DB.Create(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	// Сохраняем версию тендера
	tenderVersion := models.TenderVersion{
		TenderID:    tender.ID,
		Version:     tender.Version,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
	}
	if err := database.DB.Create(&tenderVersion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tender)
}

func GetTenders(c *gin.Context) {
	var tenders []models.Tender
	var total int64

	limit, offset := utils.GetPaginationParams(c)

	// Фильтрация по service_type
	serviceTypes := c.QueryArray("service_type")

	query := database.DB.Model(&models.Tender{}).Where("status = ?", models.TenderStatusPublished)

	if len(serviceTypes) > 0 {
		query = query.Where("service_type IN ?", serviceTypes)
	}

	query.Count(&total)

	// Пагинация и сортировка
	query.Limit(limit).Offset(offset).Order("name ASC").Find(&tenders)

	c.JSON(http.StatusOK, tenders)
}

func GetUserTenders(c *gin.Context) {
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

	var tenders []models.Tender
	var total int64

	limit, offset := utils.GetPaginationParams(c)

	// Получение тендеров, созданных пользователем
	database.DB.Model(&models.Tender{}).
		Where("organization_id IN (?)", database.DB.Model(&models.OrganizationResponsible{}).
			Select("organization_id").
			Where("user_id = ?", user.ID)).
		Count(&total).
		Limit(limit).Offset(offset).Order("name ASC").Find(&tenders)

	c.JSON(http.StatusOK, tenders)
}

func EditTender(c *gin.Context) {
	tenderID := c.Param("tenderId")
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

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}

	// Обновление тендера
	input["version"] = tender.Version + 1

	if err := database.DB.Model(&tender).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	// Сохранение версии
	tenderVersion := models.TenderVersion{
		TenderID:    tender.ID,
		Version:     tender.Version,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
	}
	if err := database.DB.Create(&tenderVersion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tender)
}

func RollbackTender(c *gin.Context) {
	tenderID := c.Param("tenderId")
	versionParam := c.Param("version")
	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Username is required"})
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

	// Получение версии тендера
	var tenderVersion models.TenderVersion
	if err := database.DB.Where("tender_id = ? AND version = ?", tender.ID, version).First(&tenderVersion).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": "Версия не найдена"})
		return
	}

	// Откат тендера
	tender.Name = tenderVersion.Name
	tender.Description = tenderVersion.Description
	tender.ServiceType = tenderVersion.ServiceType
	tender.Version += 1

	if err := database.DB.Save(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	// Сохранение новой версии
	newTenderVersion := models.TenderVersion{
		TenderID:    tender.ID,
		Version:     tender.Version,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
	}
	if err := database.DB.Create(&newTenderVersion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tender)
}

func GetTenderStatus(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{"status": tender.Status})
}

func UpdateTenderStatus(c *gin.Context) {
	tenderID := c.Param("tenderId")
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

	// Проверка корректности статуса
	validStatuses := []models.TenderStatus{
		models.TenderStatusCreated,
		models.TenderStatusPublished,
		models.TenderStatusClosed,
	}
	isValidStatus := false
	for _, s := range validStatuses {
		if s == models.TenderStatus(status) {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "Некорректный статус тендера"})
		return
	}

	// Обновление статуса тендера
	tender.Status = models.TenderStatus(status)
	if err := database.DB.Save(&tender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tender)
}

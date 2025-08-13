package handlers

import (
	"fmt"
	"net/http"

	"github.com/Aebroyx/splitbill-llmocr-api/internal/domain/models"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BillHandler struct {
	billService *services.BillService
}

func NewBillHandler(billService *services.BillService) *BillHandler {
	return &BillHandler{billService: billService}
}

// CreateBill handles bill creation
func (h *BillHandler) CreateBill(c *gin.Context) {
	var req models.BillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	bill, err := h.billService.CreateBill(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create bill: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, bill)
}

// GetBill handles retrieving a bill by ID
func (h *BillHandler) GetBill(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	bill, err := h.billService.GetBill(billID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Bill not found: %v", err)})
		return
	}

	c.JSON(http.StatusOK, bill)
}

// UploadBillImage handles image upload for a bill
func (h *BillHandler) UploadBillImage(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	// Get the uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	// Validate file type
	if !isValidImageType(file.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPG, PNG, and JPEG are allowed"})
		return
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size too large. Maximum size is 10MB"})
		return
	}

	bill, err := h.billService.UploadBillImage(billID, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload image: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image uploaded successfully",
		"bill":    bill,
	})
}

// GetBillSummary handles retrieving bill summary
func (h *BillHandler) GetBillSummary(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	summary, err := h.billService.GetBillSummary(billID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Bill not found: %v", err)})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// AddParticipant handles adding a participant to a bill
func (h *BillHandler) AddParticipant(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	var req models.ParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	participant := &models.Participants{
		BillID:             billID,
		Name:               req.Name,
		PaymentStatus:      "unpaid",
		ShareOfCommonCosts: req.ShareOfCommonCosts,
	}

	if err := h.billService.GetDB().Create(participant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add participant: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, participant)
}

// AssignItemToParticipant handles assigning an item to a participant
func (h *BillHandler) AssignItemToParticipant(c *gin.Context) {
	var req models.ItemAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	assignment := &models.ItemAssignments{
		ItemID:        req.ItemID,
		ParticipantID: req.ParticipantID,
	}

	if err := h.billService.GetDB().Create(assignment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to assign item: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, assignment)
}

// ProcessExtractedData handles processing data returned from n8n workflow
func (h *BillHandler) ProcessExtractedData(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	var req struct {
		ExtractedData string `json:"extracted_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	if err := h.billService.ProcessExtractedData(billID, req.ExtractedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process extracted data: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Extracted data processed successfully"})
}

// isValidImageType checks if the file is a valid image type
func isValidImageType(filename string) bool {
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	ext := getFileExtension(filename)
	return validExtensions[ext]
}

// getFileExtension returns the file extension in lowercase
func getFileExtension(filename string) string {
	if len(filename) == 0 {
		return ""
	}

	// Find the last dot
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}

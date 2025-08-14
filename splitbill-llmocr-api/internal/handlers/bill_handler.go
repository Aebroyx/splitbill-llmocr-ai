package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

	// Update bill status to processing
	if err := h.billService.UpdateBillStatus(billID, "processing"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update bill status: %v", err)})
		return
	}

	bill, err := h.billService.UploadBillImage(billID, file)
	if err != nil {
		// Check if it's an n8n workflow error
		if strings.Contains(err.Error(), "failed to process image with AI") {
			// Status should already be set to "failed" by the service
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to process image with AI. Please try uploading again.",
				"status":  "failed",
				"details": "The AI processing service is currently unavailable or encountered an error.",
			})
		} else {
			// Revert status to active if upload fails for other reasons
			h.billService.UpdateBillStatus(billID, "active")
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload image: %v", err)})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image uploaded successfully and sent for processing",
		"bill":    bill,
		"status":  "processing",
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

// UpdateItem handles updating an item's details
func (h *BillHandler) UpdateItem(c *gin.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req struct {
		Name     *string  `json:"name"`
		Price    *float64 `json:"price"`
		Quantity *int     `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Update only the fields that were provided
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Quantity != nil {
		updates["quantity"] = *req.Quantity
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Update the item in the database
	if err := h.billService.GetDB().Model(&models.Items{}).Where("id = ?", itemID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update item: %v", err)})
		return
	}

	// Get the updated item
	var updatedItem models.Items
	if err := h.billService.GetDB().First(&updatedItem, itemID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated item"})
		return
	}

	c.JSON(http.StatusOK, updatedItem)
}

// UpdateBill handles updating a bill's details
func (h *BillHandler) UpdateBill(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	var req struct {
		TaxAmount *float64 `json:"tax_amount"`
		TipAmount *float64 `json:"tip_amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Update only the fields that were provided
	updates := make(map[string]interface{})
	if req.TaxAmount != nil {
		updates["tax_amount"] = *req.TaxAmount
	}
	if req.TipAmount != nil {
		updates["tip_amount"] = *req.TipAmount
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Update the bill in the database
	if err := h.billService.GetDB().Model(&models.Bills{}).Where("id = ?", billID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update bill: %v", err)})
		return
	}

	// Get the updated bill
	var updatedBill models.Bills
	if err := h.billService.GetDB().First(&updatedBill, billID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated bill"})
		return
	}

	// Return the updated bill directly
	c.JSON(http.StatusOK, updatedBill)
}

// ProcessExtractedData handles processing data returned from n8n workflow
func (h *BillHandler) ProcessExtractedData(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		fmt.Printf("UUID parse error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	// Read the raw body first
	body, err := c.GetRawData()
	if err != nil {
		fmt.Printf("Error reading raw body: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	fmt.Printf("Raw request body: %s\n", string(body))

	// Parse the JSON manually since we already consumed the body
	var rawData map[string]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid JSON: %v", err)})
		return
	}

	// Declare variable for extracted data
	var extractedDataStr string

	// Check if this is the direct data structure from n8n
	if code, exists := rawData["code"]; exists && code == "API_SPLITBILL_LLMOCR" {
		fmt.Printf("Detected direct n8n data structure\n")

		// Convert the entire data to JSON string for processing
		extractedDataBytes, err := json.Marshal(rawData)
		if err != nil {
			fmt.Printf("Error marshaling data: %v\n", err)
			// Update status to failed
			h.billService.UpdateBillStatus(billID, "failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process data"})
			return
		}
		extractedDataStr = string(extractedDataBytes)
	} else {
		// Fallback: check if extracted_data field exists
		extractedData, exists := rawData["extracted_data"]
		if !exists {
			fmt.Printf("Missing extracted_data field. Available fields: %v\n", rawData)
			// Update status to failed
			h.billService.UpdateBillStatus(billID, "failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required field: extracted_data"})
			return
		}

		// Convert to string
		var ok bool
		extractedDataStr, ok = extractedData.(string)
		if !ok {
			fmt.Printf("extracted_data is not a string, it's: %T\n", extractedData)
			// Update status to failed
			h.billService.UpdateBillStatus(billID, "failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": "extracted_data must be a string"})
			return
		}
	}

	if err := h.billService.ProcessExtractedData(billID, extractedDataStr); err != nil {
		// Update status to failed
		h.billService.UpdateBillStatus(billID, "failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process extracted data: %v", err)})
		return
	}

	// Update status to completed
	if err := h.billService.UpdateBillStatus(billID, "completed"); err != nil {
		fmt.Printf("Warning: Failed to update bill status to completed: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Extracted data processed successfully"})
}

// GetBillStatus handles retrieving the status of a bill
func (h *BillHandler) GetBillStatus(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	status, err := h.billService.GetBillStatus(billID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Bill not found: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bill_id": billID,
		"status":  status,
	})
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

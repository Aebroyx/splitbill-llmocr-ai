package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"errors"

	"github.com/Aebroyx/splitbill-llmocr-api/internal/domain/models"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
		fmt.Printf("UUID parse error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	fmt.Printf("Adding participant to bill: %s\n", billID)

	var req models.ParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("JSON bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	fmt.Printf("Participant request: %+v\n", req)

	participant := &models.Participants{
		BillID:             billID,
		Name:               req.Name,
		PaymentStatus:      "unpaid",
		ShareOfCommonCosts: req.ShareOfCommonCosts,
	}

	fmt.Printf("Creating participant: %+v\n", participant)

	if err := h.billService.GetDB().Create(participant).Error; err != nil {
		fmt.Printf("Database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add participant: %v", err)})
		return
	}

	fmt.Printf("Participant created successfully with ID: %d\n", participant.ID)
	c.JSON(http.StatusCreated, participant)
}

// GetParticipants handles fetching all participants for a bill
func (h *BillHandler) GetParticipants(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		fmt.Printf("UUID parse error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	fmt.Printf("Fetching participants for bill: %s\n", billID)

	var participants []models.Participants
	if err := h.billService.GetDB().Where("bill_id = ?", billID).Find(&participants).Error; err != nil {
		fmt.Printf("Database error fetching participants: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch participants: %v", err)})
		return
	}

	fmt.Printf("Found %d participants for bill %s\n", len(participants), billID)
	c.JSON(http.StatusOK, participants)
}

// GetItemAssignments handles fetching all item assignments for a bill
func (h *BillHandler) GetItemAssignments(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		fmt.Printf("UUID parse error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	fmt.Printf("Fetching item assignments for bill: %s\n", billID)

	// Get all items for this bill
	var items []models.Items
	if err := h.billService.GetDB().Where("bill_id = ?", billID).Find(&items).Error; err != nil {
		fmt.Printf("Database error fetching items: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch items: %v", err)})
		return
	}

	fmt.Printf("Found %d items for bill %s\n", len(items), billID)
	fmt.Printf("Items: %+v\n", items)

	// Get all item assignments for these items
	var assignments []models.ItemAssignments
	if len(items) > 0 {
		itemIDs := make([]uint, len(items))
		for i, item := range items {
			itemIDs[i] = item.ID
		}

		fmt.Printf("Looking for assignments for items: %v\n", itemIDs)

		if err := h.billService.GetDB().Where("item_id IN ?", itemIDs).Find(&assignments).Error; err != nil {
			fmt.Printf("Database error fetching assignments: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch item assignments: %v", err)})
			return
		}
	} else {
		fmt.Printf("No items found for bill %s, returning empty assignments\n", billID)
	}

	fmt.Printf("Found %d item assignments for bill %s\n", len(assignments), billID)
	fmt.Printf("Assignments: %+v\n", assignments)

	c.JSON(http.StatusOK, assignments)
}

// AssignItemToParticipant handles assigning an item to a participant
func (h *BillHandler) AssignItemToParticipant(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		fmt.Printf("UUID parse error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	fmt.Printf("Assigning item to participant in bill: %s\n", billID)

	var req models.ItemAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("JSON bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	fmt.Printf("Assignment request: %+v\n", req)

	// Check if the item belongs to this bill
	var item models.Items
	if err := h.billService.GetDB().Where("id = ? AND bill_id = ?", req.ItemID, billID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Item %d not found in bill %s\n", req.ItemID, billID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in this bill"})
		} else {
			fmt.Printf("Database error finding item: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to find item: %v", err)})
		}
		return
	}

	fmt.Printf("Item found: %+v\n", item)

	// Check if the participant belongs to this bill
	var participant models.Participants
	if err := h.billService.GetDB().Where("id = ? AND bill_id = ?", req.ParticipantID, billID).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Participant %d not found in bill %s\n", req.ParticipantID, billID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Participant not found in this bill"})
		} else {
			fmt.Printf("Database error finding participant: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to find participant: %v", err)})
		}
		return
	}

	fmt.Printf("Participant found: %+v\n", participant)

	// Check if assignment already exists
	var existingAssignment models.ItemAssignments
	if err := h.billService.GetDB().Where("item_id = ? AND participant_id = ?", req.ItemID, req.ParticipantID).First(&existingAssignment).Error; err == nil {
		fmt.Printf("Assignment already exists: %+v\n", existingAssignment)
		c.JSON(http.StatusConflict, gin.H{"error": "Item is already assigned to this participant"})
		return
	}

	assignment := &models.ItemAssignments{
		ItemID:        req.ItemID,
		ParticipantID: req.ParticipantID,
	}

	fmt.Printf("Creating assignment: %+v\n", assignment)

	if err := h.billService.GetDB().Create(assignment).Error; err != nil {
		fmt.Printf("Database error creating assignment: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to assign item: %v", err)})
		return
	}

	fmt.Printf("Assignment created successfully\n")
	c.JSON(http.StatusCreated, assignment)
}

// DeleteParticipant handles deleting a participant from a bill
func (h *BillHandler) DeleteParticipant(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		fmt.Printf("UUID parse error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	participantIDStr := c.Param("participantId")
	participantID, err := strconv.ParseUint(participantIDStr, 10, 32)
	if err != nil {
		fmt.Printf("Participant ID parse error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid participant ID"})
		return
	}

	fmt.Printf("Deleting participant %d from bill %s\n", participantID, billID)

	// Check if the participant belongs to this bill
	var participant models.Participants
	if err := h.billService.GetDB().Where("id = ? AND bill_id = ?", participantID, billID).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Participant %d not found in bill %s\n", participantID, billID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Participant not found in this bill"})
		} else {
			fmt.Printf("Database error finding participant: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to find participant: %v", err)})
		}
		return
	}

	fmt.Printf("Participant found: %+v\n", participant)

	// First delete all item assignments for this participant
	fmt.Printf("Deleting item assignments for participant %d\n", participantID)
	if err := h.billService.GetDB().Where("participant_id = ?", participantID).Delete(&models.ItemAssignments{}).Error; err != nil {
		fmt.Printf("Database error deleting item assignments: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete item assignments: %v", err)})
		return
	}

	// Then delete the participant
	fmt.Printf("Deleting participant %d\n", participantID)
	if err := h.billService.GetDB().Delete(&models.Participants{}, participantID).Error; err != nil {
		fmt.Printf("Database error deleting participant: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete participant: %v", err)})
		return
	}

	fmt.Printf("Participant %d deleted successfully\n", participantID)
	c.JSON(http.StatusOK, gin.H{"message": "Participant deleted successfully"})
}

// DeleteItemAssignment handles removing an item assignment from a participant
func (h *BillHandler) DeleteItemAssignment(c *gin.Context) {
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		fmt.Printf("UUID parse error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill ID"})
		return
	}

	fmt.Printf("Deleting item assignment in bill: %s\n", billID)

	var req models.ItemAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("JSON bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	fmt.Printf("Delete assignment request: %+v\n", req)

	// Check if the item belongs to this bill
	var item models.Items
	if err := h.billService.GetDB().Where("id = ? AND bill_id = ?", req.ItemID, billID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Item %d not found in bill %s\n", req.ItemID, billID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in this bill"})
		} else {
			fmt.Printf("Database error finding item: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to find item: %v", err)})
		}
		return
	}

	fmt.Printf("Item found: %+v\n", item)

	// Check if the participant belongs to this bill
	var participant models.Participants
	if err := h.billService.GetDB().Where("id = ? AND bill_id = ?", req.ParticipantID, billID).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Participant %d not found in bill %s\n", req.ParticipantID, billID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Participant not found in this bill"})
		} else {
			fmt.Printf("Database error finding participant: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to find participant: %v", err)})
		}
		return
	}

	fmt.Printf("Participant found: %+v\n", participant)

	// Check if assignment exists
	var existingAssignment models.ItemAssignments
	if err := h.billService.GetDB().Where("item_id = ? AND participant_id = ?", req.ItemID, req.ParticipantID).First(&existingAssignment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Assignment not found for item %d and participant %d\n", req.ItemID, req.ParticipantID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Item assignment not found"})
		} else {
			fmt.Printf("Database error finding assignment: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to find assignment: %v", err)})
		}
		return
	}

	fmt.Printf("Assignment found: %+v\n", existingAssignment)

	// Delete the assignment
	if err := h.billService.GetDB().Where("item_id = ? AND participant_id = ?", req.ItemID, req.ParticipantID).Delete(&models.ItemAssignments{}).Error; err != nil {
		fmt.Printf("Database error deleting assignment: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete item assignment: %v", err)})
		return
	}

	fmt.Printf("Assignment deleted successfully\n")
	c.JSON(http.StatusOK, gin.H{"message": "Item assignment removed successfully"})
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

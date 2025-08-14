package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/Aebroyx/splitbill-llmocr-api/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillService struct {
	db *gorm.DB
}

func NewBillService(db *gorm.DB) *BillService {
	return &BillService{db: db}
}

// GetDB returns the database instance
func (s *BillService) GetDB() *gorm.DB {
	return s.db
}

// CreateBill creates a new bill
func (s *BillService) CreateBill(req *models.BillRequest) (*models.BillResponse, error) {
	bill := &models.Bills{
		ID:        uuid.New(),
		Name:      req.Name,
		Status:    "active",
		TaxAmount: req.TaxAmount,
		TipAmount: req.TipAmount,
	}

	if err := s.db.Create(bill).Error; err != nil {
		return nil, fmt.Errorf("failed to create bill: %w", err)
	}

	return s.getBillResponse(bill), nil
}

// GetBill retrieves a bill by ID
func (s *BillService) GetBill(id uuid.UUID) (*models.BillResponse, error) {
	var bill models.Bills
	if err := s.db.Preload("Items").Preload("Participants").First(&bill, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	return s.getBillResponse(&bill), nil
}

// UploadBillImage uploads an image for a bill and triggers n8n workflow
func (s *BillService) UploadBillImage(billID uuid.UUID, file *multipart.FileHeader) (*models.BillResponse, error) {
	// Check if bill exists
	bill, err := s.GetBill(billID)
	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	// Read file data
	fileBytes, err := s.readFileData(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	// Save image to disk (optional, for backup)
	imagePath := fmt.Sprintf("./uploads/bill_%s_%s", billID.String(), file.Filename)
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
		// Don't fail the upload for this, continue with n8n
	}

	if err := os.WriteFile(imagePath, fileBytes, 0644); err != nil {
		fmt.Printf("Failed to save image to disk: %v\n", err)
		// Don't fail the upload for this, continue with n8n
	}

	// Trigger n8n workflow with image data
	if err := s.triggerN8nWorkflowWithImage(billID, fileBytes, file.Filename); err != nil {
		// If n8n workflow fails, the status should already be set to "failed"
		// but let's make sure we return a proper error message
		fmt.Printf("N8n workflow failed for bill %s: %v\n", billID, err)
		return nil, fmt.Errorf("failed to process image with AI: %w", err)
	}

	return bill, nil
}

// readFileData reads the file data from multipart.FileHeader into bytes
func (s *BillService) readFileData(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read file content into bytes
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return fileBytes, nil
}

// triggerN8nWorkflowWithImage sends the image data directly to n8n workflow
func (s *BillService) triggerN8nWorkflowWithImage(billID uuid.UUID, imageData []byte, filename string) error {
	n8nWebhookURL := os.Getenv("N8N_WEBHOOK_URL")
	if n8nWebhookURL == "" {
		err := fmt.Errorf("N8N_WEBHOOK_URL not configured")
		fmt.Printf("N8N_WEBHOOK_URL not configured, skipping workflow trigger for bill %s\n", billID)
		// Update bill status to failed since we can't process
		if updateErr := s.UpdateBillStatus(billID, "failed"); updateErr != nil {
			fmt.Printf("Failed to update bill status to failed: %v\n", updateErr)
		}
		return err
	}

	// Create multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add bill_id field
	if err := writer.WriteField("bill_id", billID.String()); err != nil {
		fmt.Printf("Failed to write bill_id field: %v\n", err)
		// Update bill status to failed
		if updateErr := s.UpdateBillStatus(billID, "failed"); updateErr != nil {
			fmt.Printf("Failed to update bill status to failed: %v\n", updateErr)
		}
		return fmt.Errorf("failed to write bill_id field: %v", err)
	}

	// Add image file
	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		fmt.Printf("Failed to create form file: %v\n", err)
		// Update bill status to failed
		if updateErr := s.UpdateBillStatus(billID, "failed"); updateErr != nil {
			fmt.Printf("Failed to update bill status to failed: %v\n", updateErr)
		}
		return fmt.Errorf("failed to create form file: %v", err)
	}
	if _, err := part.Write(imageData); err != nil {
		fmt.Printf("Failed to write image data: %v\n", err)
		// Update bill status to failed
		if updateErr := s.UpdateBillStatus(billID, "failed"); updateErr != nil {
			fmt.Printf("Failed to update bill status to failed: %v\n", updateErr)
		}
		return fmt.Errorf("failed to write image data: %v", err)
	}

	// Get the Content-Type BEFORE closing the writer
	contentType := writer.FormDataContentType()

	// Close the writer to finalize the multipart data
	writer.Close()

	// Send request to n8n
	req, err := http.NewRequest("POST", n8nWebhookURL, &requestBody)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		// Update bill status to failed
		if updateErr := s.UpdateBillStatus(billID, "failed"); updateErr != nil {
			fmt.Printf("Failed to update bill status to failed: %v\n", updateErr)
		}
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the Content-Type header with the boundary
	req.Header.Set("Content-Type", contentType)

	// Set timeout for the request
	client := &http.Client{
		Timeout: 30 * time.Second, // 30 second timeout
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request to n8n: %v\n", err)
		// Update bill status to failed
		if updateErr := s.UpdateBillStatus(billID, "failed"); updateErr != nil {
			fmt.Printf("Failed to update bill status to failed: %v\n", updateErr)
		}
		return fmt.Errorf("failed to send request to n8n: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("N8n workflow returned status: %d\n", resp.StatusCode)
		fmt.Printf("Response body: %s\n", string(bodyBytes))
		fmt.Printf("Request headers: %v\n", req.Header)

		// Update bill status to failed since n8n workflow failed
		if updateErr := s.UpdateBillStatus(billID, "failed"); updateErr != nil {
			fmt.Printf("Failed to update bill status to failed: %v\n", updateErr)
		}

		return fmt.Errorf("n8n workflow failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Printf("Successfully triggered n8n workflow for bill %s\n", billID)
	return nil
}

// ProcessExtractedData processes the data returned from n8n workflow
func (s *BillService) ProcessExtractedData(billID uuid.UUID, extractedData string) error {
	var bill models.Bills
	if err := s.db.First(&bill, "id = ?", billID).Error; err != nil {
		return fmt.Errorf("bill not found: %w", err)
	}

	// Parse the extracted data
	var extractedItems models.ExtractedItemData
	if err := json.Unmarshal([]byte(extractedData), &extractedItems); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		return fmt.Errorf("failed to parse extracted data: %w", err)
	}

	// Start a transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update bill with extracted data (only tax and tip amounts)
	if err := tx.Model(&bill).Updates(map[string]interface{}{
		"tax_amount": extractedItems.Tax,
		"tip_amount": extractedItems.Tip,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update bill: %w", err)
	}

	// Create items from extracted data
	for _, item := range extractedItems.Items {
		dbItem := models.Items{
			BillID:   billID,
			Name:     item.Name,
			Price:    item.Price,
			Quantity: item.Quantity,
		}

		if err := tx.Create(&dbItem).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create item: %w", err)
		}
	}

	return tx.Commit().Error
}

// GetBillSummary calculates and returns bill summary
func (s *BillService) GetBillSummary(billID uuid.UUID) (*models.BillSummary, error) {
	var bill models.Bills
	if err := s.db.Preload("Items").Preload("Participants").First(&bill, "id = ?", billID).Error; err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	// Calculate total items
	var totalItems float64
	for _, item := range bill.Items {
		totalItems += item.Price * float64(item.Quantity)
	}

	// Calculate participant shares
	participantShares := make(map[string]float64)
	totalParticipants := len(bill.Participants)
	if totalParticipants > 0 {
		sharePerPerson := (totalItems + bill.TaxAmount + bill.TipAmount) / float64(totalParticipants)
		for _, participant := range bill.Participants {
			participantShares[participant.Name] = sharePerPerson + participant.ShareOfCommonCosts
		}
	}

	return &models.BillSummary{
		BillID:            billID,
		TotalItems:        totalItems,
		TaxAmount:         bill.TaxAmount,
		TipAmount:         bill.TipAmount,
		TotalBill:         totalItems + bill.TaxAmount + bill.TipAmount,
		ParticipantShares: participantShares,
	}, nil
}

// UpdateBillStatus updates the status of a bill
func (s *BillService) UpdateBillStatus(billID uuid.UUID, status string) error {
	return s.db.Model(&models.Bills{}).Where("id = ?", billID).Update("status", status).Error
}

// GetBillStatus returns the current status of a bill
func (s *BillService) GetBillStatus(billID uuid.UUID) (string, error) {
	var bill models.Bills
	err := s.db.Select("status").Where("id = ?", billID).First(&bill).Error
	if err != nil {
		return "", err
	}
	return bill.Status, nil
}

// getBillResponse converts a Bills model to BillResponse
func (s *BillService) getBillResponse(bill *models.Bills) *models.BillResponse {
	response := &models.BillResponse{
		ID:        bill.ID,
		Name:      bill.Name,
		Status:    bill.Status,
		TaxAmount: bill.TaxAmount,
		TipAmount: bill.TipAmount,
		CreatedAt: bill.CreatedAt,
	}

	// Convert items
	for _, item := range bill.Items {
		response.Items = append(response.Items, models.ItemResponse{
			ID:        item.ID,
			BillID:    item.BillID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
			CreatedAt: item.CreatedAt,
		})
	}

	// Convert participants
	for _, participant := range bill.Participants {
		response.Participants = append(response.Participants, models.ParticipantResponse{
			ID:                 participant.ID,
			BillID:             participant.BillID,
			Name:               participant.Name,
			PaymentStatus:      participant.PaymentStatus,
			ShareOfCommonCosts: participant.ShareOfCommonCosts,
			CreatedAt:          participant.CreatedAt,
		})
	}

	return response
}

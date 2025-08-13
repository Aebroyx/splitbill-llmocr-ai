# Split Bill LLM OCR API

A Go-based API for splitting bills using image processing and LLM (Gemini) integration through n8n workflows.

## Features

- Upload bill images for processing
- Automatic bill data extraction using Gemini LLM
- Bill splitting and participant management
- Item assignment to participants
- Bill summary calculations

## API Endpoints

### Bills

#### Create a new bill
```
POST /api/bills/
Content-Type: application/json

{
  "name": "Restaurant Bill",
  "tax_amount": 5.50,
  "tip_amount": 10.00
}
```

#### Get bill by ID
```
GET /api/bills/{id}
```

#### Upload bill image
```
POST /api/bills/{id}/image
Content-Type: multipart/form-data

Form data:
- image: [image file] (JPG, PNG, JPEG, max 10MB)
```

#### Get bill summary
```
GET /api/bills/{id}/summary
```

#### Add participant to bill
```
POST /api/bills/{id}/participants
Content-Type: application/json

{
  "name": "John Doe",
  "share_of_common_costs": 2.50
}
```

#### Assign item to participant
```
POST /api/bills/{id}/assign-items
Content-Type: application/json

{
  "item_id": 1,
  "participant_id": 1
}
```

#### Process extracted data (for n8n workflow)
```
POST /api/bills/{id}/process-data
Content-Type: application/json

{
  "extracted_data": "{\"items\":[{\"name\":\"Burger\",\"price\":12.99,\"quantity\":1}],\"tax\":1.30,\"tip\":2.60,\"total\":16.89}"
}
```

## Environment Variables

Create a `.env` file in the root directory:

```env
# Server
SERVER_PORT=8080
SERVER_HOST=localhost

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=splitbill-llmocr.app
DB_SSL_MODE=disable

# JWT (if needed for other features)
JWT_SECRET=your_jwt_secret
JWT_EXPIRY=24h

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000

# N8N Integration
N8N_WEBHOOK_URL=http://localhost:5678/webhook/bill-processing
```

## Setup

1. Install dependencies:
```bash
go mod tidy
```

2. Set up PostgreSQL database and update the `.env` file

3. Run the application:
```bash
go run cmd/main.go
```

## N8N Workflow Integration

The API integrates with n8n workflows for image processing:

1. When an image is uploaded, the API sends a webhook to the configured n8n workflow
2. The n8n workflow should:
   - Receive the image
   - Process it using Gemini LLM
   - Extract bill items, tax, tip, and total
   - Return the structured data to the API

### Expected n8n workflow payload:
```json
{
  "bill_id": "uuid-string",
  "image_path": "/path/to/image.jpg",
  "timestamp": 1234567890
}
```

### Expected n8n workflow response:
```json
{
  "extracted_data": "{\"items\":[{\"name\":\"Item Name\",\"price\":10.99,\"quantity\":1}],\"tax\":1.10,\"tip\":2.20,\"total\":14.29}"
}
```

## File Structure

```
splitbill-llmocr-api/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go          # Configuration management
│   ├── database/
│   │   └── db.go              # Database connection
│   ├── domain/
│   │   └── models/
│   │       ├── bills.go       # Bill-related models
│   │       └── users.go       # User models
│   ├── handlers/
│   │   ├── auth_handler.go    # Authentication handlers
│   │   └── bill_handler.go    # Bill-related handlers
│   ├── middleware/
│   │   └── auth.go            # Authentication middleware
│   └── services/
│       ├── user_service.go    # User business logic
│       └── bill_service.go    # Bill business logic
├── uploads/                   # Uploaded images directory
├── go.mod
├── go.sum
└── README.md
```

## Usage Example

1. Create a bill:
```bash
curl -X POST http://localhost:8080/api/bills/ \
  -H "Content-Type: application/json" \
  -d '{"name": "Dinner Bill", "tax_amount": 5.00, "tip_amount": 10.00}'
```

2. Upload an image:
```bash
curl -X POST http://localhost:8080/api/bills/{bill-id}/image \
  -F "image=@/path/to/bill-image.jpg"
```

3. Get bill summary:
```bash
curl http://localhost:8080/api/bills/{bill-id}/summary
```

## Notes

- This is an open API - no authentication required for bill operations
- Images are stored locally in the `uploads/` directory
- The API automatically triggers n8n workflows when images are uploaded
- All monetary values are stored as decimal numbers with 2 decimal places


package test

import (
	"context"
	"testing"
	"time"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

func TestBatchQRCodeService_CreateBatchQRCode(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.BatchQRCode{}, &domain.BatchQRCodeItem{}, &domain.Application{})

	batchQRCodeService := service.NewBatchQRCodeServiceWithDB(db)

	// Create a test application first
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "qrcode",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		request       *service.CreateBatchQRCodeRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid static batch QR code creation",
			request: &service.CreateBatchQRCodeRequest{
				AppID:       "app-001",
				Name:        "Static Batch",
				Description: "Static batch QR codes",
				Prefix:      "STATIC",
				Count:       10,
				Type:        "static",
				URLTemplate: "https://example.com/static/{id}",
				CreatedBy:   "user-001",
			},
			expectError: false,
		},
		{
			name: "Valid dynamic batch QR code creation",
			request: &service.CreateBatchQRCodeRequest{
				AppID:       "app-001",
				Name:        "Dynamic Batch",
				Description: "Dynamic batch QR codes",
				Prefix:      "DYNAMIC",
				Count:       5,
				Type:        "dynamic",
				URLTemplate: "https://example.com/dynamic/{id}",
				CreatedBy:   "user-001",
			},
			expectError: false,
		},
		{
			name: "Invalid QR code type",
			request: &service.CreateBatchQRCodeRequest{
				AppID:       "app-001",
				Name:        "Invalid Type Batch",
				Description: "Invalid type batch QR codes",
				Prefix:      "INVALID",
				Count:       10,
				Type:        "invalid",
				URLTemplate: "https://example.com/invalid/{id}",
				CreatedBy:   "user-001",
			},
			expectError:  true,
			errorMessage: "invalid QR code type",
		},
		{
			name: "Invalid count - too low",
			request: &service.CreateBatchQRCodeRequest{
				AppID:       "app-001",
				Name:        "Low Count Batch",
				Description: "Low count batch QR codes",
				Prefix:      "LOW",
				Count:       0,
				Type:        "static",
				URLTemplate: "https://example.com/low/{id}",
				CreatedBy:   "user-001",
			},
			expectError:  true,
			errorMessage: "invalid count, must be between 1 and 10000",
		},
		{
			name: "Invalid count - too high",
			request: &service.CreateBatchQRCodeRequest{
				AppID:       "app-001",
				Name:        "High Count Batch",
				Description: "High count batch QR codes",
				Prefix:      "HIGH",
				Count:       10001,
				Type:        "static",
				URLTemplate: "https://example.com/high/{id}",
				CreatedBy:   "user-001",
			},
			expectError:  true,
			errorMessage: "invalid count, must be between 1 and 10000",
		},
		{
			name: "Non-existent application",
			request: &service.CreateBatchQRCodeRequest{
				AppID:       "non-existent-app",
				Name:        "Orphan Batch",
				Description: "Orphan batch QR codes",
				Prefix:      "ORPHAN",
				Count:       10,
				Type:        "static",
				URLTemplate: "https://example.com/orphan/{id}",
				CreatedBy:   "user-001",
			},
			expectError: false, // Note: This doesn't validate app existence in the service
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			batch, err := batchQRCodeService.CreateBatchQRCode(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, batch)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, batch)
				assert.NotEmpty(t, batch.ID)
				assert.Equal(t, tt.request.AppID, batch.AppID)
				assert.Equal(t, tt.request.Name, batch.Name)
				assert.Equal(t, tt.request.Description, batch.Description)
				assert.Equal(t, tt.request.Prefix, batch.Prefix)
				assert.Equal(t, tt.request.Count, batch.Count)
				assert.Equal(t, tt.request.Type, batch.Type)
				assert.Equal(t, tt.request.URLTemplate, batch.URLTemplate)
				assert.Equal(t, "pending", batch.Status)
				assert.Equal(t, tt.request.CreatedBy, batch.CreatedBy)
				assert.WithinDuration(t, time.Now(), batch.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), batch.UpdatedAt, time.Second)
			}
		})
	}
}

func TestBatchQRCodeService_UpdateBatchQRCode(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.BatchQRCode{}, &domain.BatchQRCodeItem{}, &domain.Application{})

	batchQRCodeService := service.NewBatchQRCodeServiceWithDB(db)

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "qrcode",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// We need to use the service to create the batch QR code to ensure it's properly initialized
	createReq := &service.CreateBatchQRCodeRequest{
		AppID:       "app-001",
		Name:        "Original Batch",
		Description: "Original description",
		Prefix:      "ORIGINAL",
		Count:       10,
		Type:        "static",
		URLTemplate: "https://example.com/original/{id}",
		CreatedBy:   "user-001",
	}
	createdBatch, err := batchQRCodeService.CreateBatchQRCode(context.Background(), createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdBatch)

	// Test cases
	tests := []struct {
		name          string
		batchID       string
		request       *service.UpdateBatchQRCodeRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name:    "Valid batch QR code update",
			batchID: createdBatch.ID,
			request: &service.UpdateBatchQRCodeRequest{
				Name:        "Updated Batch",
				Description: "Updated description",
				Prefix:      "UPDATED",
				URLTemplate: "https://example.com/updated/{id}",
				Config:      `{"color": "blue"}`,
			},
			expectError: false,
		},
		{
			name:          "Update non-existent batch QR code",
			batchID:       "non-existent-id",
			request:       &service.UpdateBatchQRCodeRequest{Name: "Updated Name"},
			expectError:   true,
			errorMessage:  "batch QR code not found",
		},
		{
			name:    "Partial update - name only",
			batchID: createdBatch.ID,
			request: &service.UpdateBatchQRCodeRequest{
				Name: "Name Only Update",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := batchQRCodeService.UpdateBatchQRCode(context.Background(), tt.batchID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify the update
				updatedBatch, getErr := batchQRCodeService.GetBatchQRCode(context.Background(), tt.batchID)
				assert.NoError(t, getErr)
				assert.NotNil(t, updatedBatch)

				// Check updated fields
				if tt.request.Name != "" {
					assert.Equal(t, tt.request.Name, updatedBatch.Name)
				}
				if tt.request.Description != "" {
					assert.Equal(t, tt.request.Description, updatedBatch.Description)
				}
				if tt.request.Prefix != "" {
					assert.Equal(t, tt.request.Prefix, updatedBatch.Prefix)
				}
				if tt.request.URLTemplate != "" {
					assert.Equal(t, tt.request.URLTemplate, updatedBatch.URLTemplate)
				}
				if tt.request.Config != "" {
					assert.Equal(t, tt.request.Config, updatedBatch.Config)
				}
				// UpdatedAt should be changed
				assert.True(t, updatedBatch.UpdatedAt.After(createdBatch.UpdatedAt))
			}
		})
	}
}

func TestBatchQRCodeService_DeleteBatchQRCode(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.BatchQRCode{}, &domain.BatchQRCodeItem{}, &domain.Application{})

	batchQRCodeService := service.NewBatchQRCodeServiceWithDB(db)

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "qrcode",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a batch QR code for testing
	createReq := &service.CreateBatchQRCodeRequest{
		AppID:       "app-001",
		Name:        "Test Batch",
		Description: "Test description",
		Prefix:      "TEST",
		Count:       5,
		Type:        "static",
		URLTemplate: "https://example.com/test/{id}",
		CreatedBy:   "user-001",
	}
	createdBatch, err := batchQRCodeService.CreateBatchQRCode(context.Background(), createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdBatch)

	// Test cases
	tests := []struct {
		name          string
		batchID       string
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "Delete non-existent batch QR code",
			batchID:     "non-existent-id",
			expectError: true,
			errorMessage: "batch QR code not found",
		},
		{
			name:        "Valid batch QR code deletion",
			batchID:     createdBatch.ID,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := batchQRCodeService.DeleteBatchQRCode(context.Background(), tt.batchID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify batch QR code is deleted
				_, getErr := batchQRCodeService.GetBatchQRCode(context.Background(), tt.batchID)
				assert.Error(t, getErr)
				assert.Equal(t, "batch QR code not found", getErr.Error())
			}
		})
	}
}

func TestBatchQRCodeService_ListBatchQRCodes(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.BatchQRCode{}, &domain.BatchQRCodeItem{}, &domain.Application{})

	batchQRCodeService := service.NewBatchQRCodeServiceWithDB(db)

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "qrcode",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create test batch QR codes
	batches := []*service.CreateBatchQRCodeRequest{
		{
			AppID:       "app-001",
			Name:        "Batch 1",
			Description: "First batch",
			Prefix:      "BATCH1",
			Count:       10,
			Type:        "static",
			URLTemplate: "https://example.com/batch1/{id}",
			CreatedBy:   "user-001",
		},
		{
			AppID:       "app-001",
			Name:        "Batch 2",
			Description: "Second batch",
			Prefix:      "BATCH2",
			Count:       5,
			Type:        "dynamic",
			URLTemplate: "https://example.com/batch2/{id}",
			CreatedBy:   "user-001",
		},
		{
			AppID:       "app-001",
			Name:        "Batch 3",
			Description: "Third batch",
			Prefix:      "BATCH3",
			Count:       15,
			Type:        "static",
			URLTemplate: "https://example.com/batch3/{id}",
			CreatedBy:   "user-001",
		},
	}

	for _, batchReq := range batches {
		_, err := batchQRCodeService.CreateBatchQRCode(context.Background(), batchReq)
		assert.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name              string
		appID             string
		page              int
		size              int
		expectedCount     int
		totalCount        int64
		expectError       bool
	}{
		{
			name:          "List first page",
			appID:         "app-001",
			page:          1,
			size:          2,
			expectedCount: 2,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List second page",
			appID:         "app-001",
			page:          2,
			size:          2,
			expectedCount: 1,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List with large page size",
			appID:         "app-001",
			page:          1,
			size:          10,
			expectedCount: 3,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List with zero page",
			appID:         "app-001",
			page:          0,
			size:          10,
			expectedCount: 3,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List non-existent app",
			appID:         "non-existent-app",
			page:          1,
			size:          10,
			expectedCount: 0,
			totalCount:    0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, total, err := batchQRCodeService.ListBatchQRCodes(context.Background(), tt.appID, tt.page, tt.size)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(result))
				assert.Equal(t, tt.totalCount, total)

				// Verify ordering (should be by created_at desc)
				if len(result) > 1 {
					for i := 0; i < len(result)-1; i++ {
						assert.True(t, result[i].CreatedAt.After(result[i+1].CreatedAt) || 
							result[i].CreatedAt.Equal(result[i+1].CreatedAt))
					}
				}
			}
		})
	}
}

func TestBatchQRCodeService_GetBatchQRCode(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.BatchQRCode{}, &domain.BatchQRCodeItem{}, &domain.Application{})

	batchQRCodeService := service.NewBatchQRCodeServiceWithDB(db)

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "qrcode",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a batch QR code for testing
	createReq := &service.CreateBatchQRCodeRequest{
		AppID:       "app-001",
		Name:        "Test Batch",
		Description: "Test description",
		Prefix:      "TEST",
		Count:       5,
		Type:        "static",
		URLTemplate: "https://example.com/test/{id}",
		CreatedBy:   "user-001",
	}
	createdBatch, err := batchQRCodeService.CreateBatchQRCode(context.Background(), createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdBatch)

	// Test cases
	tests := []struct {
		name          string
		batchID       string
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "Get non-existent batch QR code",
			batchID:     "non-existent-id",
			expectError: true,
			errorMessage: "batch QR code not found",
		},
		{
			name:        "Get existing batch QR code",
			batchID:     createdBatch.ID,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := batchQRCodeService.GetBatchQRCode(context.Background(), tt.batchID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, createdBatch.ID, result.ID)
				assert.Equal(t, createdBatch.AppID, result.AppID)
				assert.Equal(t, createdBatch.Name, result.Name)
				assert.Equal(t, createdBatch.Description, result.Description)
				assert.Equal(t, createdBatch.Prefix, result.Prefix)
				assert.Equal(t, createdBatch.Count, result.Count)
				assert.Equal(t, createdBatch.Type, result.Type)
				assert.Equal(t, createdBatch.URLTemplate, result.URLTemplate)
				assert.Equal(t, createdBatch.Status, result.Status)
				assert.Equal(t, createdBatch.CreatedBy, result.CreatedBy)
				assert.Equal(t, createdBatch.CreatedAt.Unix(), result.CreatedAt.Unix())
				assert.Equal(t, createdBatch.UpdatedAt.Unix(), result.UpdatedAt.Unix())
			}
		})
	}
}
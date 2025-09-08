package qrcode_test

import (
	"context"
	"testing"
	"time"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

func TestQRCodeService_CreateQRCode(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.QRCode{}, &domain.Application{})

	qrCodeService := service.NewQRCodeService()

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
		request       *service.CreateQRCodeRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid QR code creation",
			request: &service.CreateQRCodeRequest{
				AppID:     "app-001",
				Name:      "Test QR Code",
				Content:   "https://example.com",
				Type:      "static",
				URL:       "https://example.com",
				CreatedBy: "user-001",
			},
			expectError: false,
		},
		{
			name: "Invalid QR code type",
			request: &service.CreateQRCodeRequest{
				AppID:     "app-001",
				Name:      "Invalid QR Code",
				Content:   "https://example.com",
				Type:      "invalid",
				URL:       "https://example.com",
				CreatedBy: "user-001",
			},
			expectError:  true,
			errorMessage: "invalid QR code type",
		},
		{
			name: "Non-existent application",
			request: &service.CreateQRCodeRequest{
				AppID:     "non-existent-app",
				Name:      "Orphan QR Code",
				Content:   "https://example.com",
				Type:      "static",
				URL:       "https://example.com",
				CreatedBy: "user-001",
			},
			expectError: false, // Note: This doesn't validate app existence in the service
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			qrCode, err := qrCodeService.CreateQRCode(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, qrCode)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, qrCode)
				assert.NotEmpty(t, qrCode.ID)
				assert.Equal(t, tt.request.AppID, qrCode.AppID)
				assert.Equal(t, tt.request.Name, qrCode.Name)
				assert.Equal(t, tt.request.Content, qrCode.Content)
				assert.Equal(t, tt.request.Type, qrCode.Type)
				assert.Equal(t, tt.request.URL, qrCode.URL)
				assert.Equal(t, tt.request.CreatedBy, qrCode.CreatedBy)
				assert.WithinDuration(t, time.Now(), qrCode.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), qrCode.UpdatedAt, time.Second)
			}
		})
	}
}

func TestQRCodeService_UpdateQRCode(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.QRCode{}, &domain.Application{})

	qrCodeService := service.NewQRCodeService()

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

	// Create a QR code for testing
	qrCode := &domain.QRCode{
		ID:        "qrcode-001",
		AppID:     "app-001",
		Name:      "Original QR Code",
		Content:   "https://example.com/original",
		Type:      "static",
		URL:       "https://example.com/original",
		CreatedBy: "user-001",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	err = db.Create(qrCode).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		qrCodeID      string
		request       *service.UpdateQRCodeRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name:     "Valid QR code update",
			qrCodeID: "qrcode-001",
			request: &service.UpdateQRCodeRequest{
				Name:    "Updated QR Code",
				Content: "https://example.com/updated",
				URL:     "https://example.com/updated",
			},
			expectError: false,
		},
		{
			name:        "Update non-existent QR code",
			qrCodeID:    "non-existent-id",
			request:     &service.UpdateQRCodeRequest{Name: "Updated Name"},
			expectError: true,
			errorMessage: "QR code not found",
		},
		{
			name:     "Partial update - name only",
			qrCodeID: "qrcode-001",
			request: &service.UpdateQRCodeRequest{
				Name: "Name Only Update",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := qrCodeService.UpdateQRCode(context.Background(), tt.qrCodeID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify the update
				updatedQRCode, getErr := qrCodeService.GetQRCode(context.Background(), tt.qrCodeID)
				assert.NoError(t, getErr)
				assert.NotNil(t, updatedQRCode)

				// Check updated fields
				if tt.request.Name != "" {
					assert.Equal(t, tt.request.Name, updatedQRCode.Name)
				}
				if tt.request.Content != "" {
					assert.Equal(t, tt.request.Content, updatedQRCode.Content)
				}
				if tt.request.URL != "" {
					assert.Equal(t, tt.request.URL, updatedQRCode.URL)
				}
				// UpdatedAt should be changed
				assert.True(t, updatedQRCode.UpdatedAt.After(qrCode.UpdatedAt))
			}
		})
	}
}

func TestQRCodeService_DeleteQRCode(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.QRCode{}, &domain.Application{})

	qrCodeService := service.NewQRCodeService()

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

	// Create a QR code for testing
	qrCode := &domain.QRCode{
		ID:        "qrcode-001",
		AppID:     "app-001",
		Name:      "Test QR Code",
		Content:   "https://example.com",
		Type:      "static",
		URL:       "https://example.com",
		CreatedBy: "user-001",
	}
	err = db.Create(qrCode).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		qrCodeID      string
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "Delete non-existent QR code",
			qrCodeID:    "non-existent-id",
			expectError: true,
			errorMessage: "QR code not found",
		},
		{
			name:        "Valid QR code deletion",
			qrCodeID:    "qrcode-001",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := qrCodeService.DeleteQRCode(context.Background(), tt.qrCodeID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify QR code is deleted
				_, getErr := qrCodeService.GetQRCode(context.Background(), tt.qrCodeID)
				assert.Error(t, getErr)
				assert.Equal(t, "QR code not found", getErr.Error())
			}
		})
	}
}

func TestQRCodeService_ListQRCodes(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.QRCode{}, &domain.Application{})

	qrCodeService := service.NewQRCodeService()

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

	// Create test QR codes
	qrCodes := []*domain.QRCode{
		{
			ID:        "qrcode-001",
			AppID:     "app-001",
			Name:      "QR Code 1",
			Content:   "https://example.com/1",
			Type:      "static",
			URL:       "https://example.com/1",
			CreatedBy: "user-001",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "qrcode-002",
			AppID:     "app-001",
			Name:      "QR Code 2",
			Content:   "https://example.com/2",
			Type:      "dynamic",
			URL:       "https://example.com/2",
			CreatedBy: "user-001",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        "qrcode-003",
			AppID:     "app-001",
			Name:      "QR Code 3",
			Content:   "https://example.com/3",
			Type:      "static",
			URL:       "https://example.com/3",
			CreatedBy: "user-001",
			CreatedAt: time.Now(),
		},
	}

	for _, qrCode := range qrCodes {
		err := db.Create(qrCode).Error
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
			result, total, err := qrCodeService.ListQRCodes(context.Background(), tt.appID, tt.page, tt.size)

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

func TestQRCodeService_GetQRCode(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.QRCode{}, &domain.Application{})

	qrCodeService := service.NewQRCodeService()

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

	// Create a QR code for testing
	qrCode := &domain.QRCode{
		ID:        "qrcode-001",
		AppID:     "app-001",
		Name:      "Test QR Code",
		Content:   "https://example.com",
		Type:      "static",
		URL:       "https://example.com",
		CreatedBy: "user-001",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	err = db.Create(qrCode).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		qrCodeID      string
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "Get non-existent QR code",
			qrCodeID:    "non-existent-id",
			expectError: true,
			errorMessage: "QR code not found",
		},
		{
			name:        "Get existing QR code",
			qrCodeID:    "qrcode-001",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := qrCodeService.GetQRCode(context.Background(), tt.qrCodeID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, qrCode.ID, result.ID)
				assert.Equal(t, qrCode.AppID, result.AppID)
				assert.Equal(t, qrCode.Name, result.Name)
				assert.Equal(t, qrCode.Content, result.Content)
				assert.Equal(t, qrCode.Type, result.Type)
				assert.Equal(t, qrCode.URL, result.URL)
				assert.Equal(t, qrCode.CreatedBy, result.CreatedBy)
				assert.Equal(t, qrCode.CreatedAt.Unix(), result.CreatedAt.Unix())
				assert.Equal(t, qrCode.UpdatedAt.Unix(), result.UpdatedAt.Unix())
			}
		})
	}
}

func TestQRCodeService_GenerateQRCodeImage(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.QRCode{}, &domain.Application{})

	qrCodeService := service.NewQRCodeService()

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

	// Create a QR code for testing
	qrCode := &domain.QRCode{
		ID:        "qrcode-001",
		AppID:     "app-001",
		Name:      "Test QR Code",
		Content:   "https://example.com",
		Type:      "static",
		URL:       "https://example.com",
		CreatedBy: "user-001",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	err = db.Create(qrCode).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		qrCodeID      string
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "Generate image for non-existent QR code",
			qrCodeID:    "non-existent-id",
			expectError: true,
			errorMessage: "QR code not found",
		},
		{
			name:        "Generate image for existing QR code",
			qrCodeID:    "qrcode-001",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			imagePath, err := qrCodeService.GenerateQRCodeImage(context.Background(), tt.qrCodeID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Empty(t, imagePath)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, imagePath)
			}
		})
	}
}
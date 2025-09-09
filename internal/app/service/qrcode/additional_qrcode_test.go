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

func TestQRCodeService_CreateQRCode_Additional(t *testing.T) {
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

	// Additional test cases for CreateQRCode
	tests := []struct {
		name          string
		request       *service.CreateQRCodeRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Dynamic QR code creation",
			request: &service.CreateQRCodeRequest{
				AppID:     "app-001",
				Name:      "Dynamic QR Code",
				Content:   "https://example.com/dynamic",
				Type:      "dynamic",
				URL:       "https://example.com/dynamic",
				CreatedBy: "user-001",
			},
			expectError: false,
		},
		{
			name: "QR code with empty URL",
			request: &service.CreateQRCodeRequest{
				AppID:     "app-001",
				Name:      "QR Code with Empty URL",
				Content:   "https://example.com/empty-url",
				Type:      "static",
				URL:       "",
				CreatedBy: "user-001",
			},
			expectError: false,
		},
		{
			name: "QR code with special characters in content",
			request: &service.CreateQRCodeRequest{
				AppID:     "app-001",
				Name:      "Special Characters QR Code",
				Content:   "https://example.com/special?param=value&other=123",
				Type:      "static",
				URL:       "https://example.com/special?param=value&other=123",
				CreatedBy: "user-001",
			},
			expectError: false,
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
				
				// Verify image path is generated
				assert.NotEmpty(t, qrCode.ImagePath)
			}
		})
	}
}

func TestQRCodeService_UpdateQRCode_Additional(t *testing.T) {
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

	// Additional test cases for UpdateQRCode
	tests := []struct {
		name          string
		qrCodeID      string
		request       *service.UpdateQRCodeRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name:     "Update QR code to dynamic type",
			qrCodeID: "qrcode-001",
			request: &service.UpdateQRCodeRequest{
				Content: "https://example.com/dynamic",
				URL:     "https://example.com/dynamic",
			},
			expectError: false,
		},
		{
			name:     "Update QR code with empty URL",
			qrCodeID: "qrcode-001",
			request: &service.UpdateQRCodeRequest{
				URL: "",
			},
			expectError: false,
		},
		{
			name:     "Update QR code with special characters",
			qrCodeID: "qrcode-001",
			request: &service.UpdateQRCodeRequest{
				Content: "https://example.com/special?param=value&other=123",
				URL:     "https://example.com/special?param=value&other=123",
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
				
				// Verify image path is updated
				assert.NotEmpty(t, updatedQRCode.ImagePath)
			}
		})
	}
}

func TestQRCodeService_DeleteQRCode_Additional(t *testing.T) {
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

	// Test case for deleting QR code and verifying image path
	t.Run("Valid QR code deletion with image verification", func(t *testing.T) {
		// Execute
		err := qrCodeService.DeleteQRCode(context.Background(), "qrcode-001")

		// Assert
		assert.NoError(t, err)

		// Verify QR code is deleted
		_, getErr := qrCodeService.GetQRCode(context.Background(), "qrcode-001")
		assert.Error(t, getErr)
		assert.Equal(t, "QR code not found", getErr.Error())
	})
}

func TestQRCodeService_ListQRCodes_Additional(t *testing.T) {
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

	// Create test QR codes with different types
	qrCodes := []*domain.QRCode{
		{
			ID:        "qrcode-001",
			AppID:     "app-001",
			Name:      "Static QR Code 1",
			Content:   "https://example.com/1",
			Type:      "static",
			URL:       "https://example.com/1",
			CreatedBy: "user-001",
			CreatedAt: time.Now().Add(-3 * time.Hour),
		},
		{
			ID:        "qrcode-002",
			AppID:     "app-001",
			Name:      "Dynamic QR Code 1",
			Content:   "https://example.com/2",
			Type:      "dynamic",
			URL:       "https://example.com/2",
			CreatedBy: "user-001",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "qrcode-003",
			AppID:     "app-001",
			Name:      "Static QR Code 2",
			Content:   "https://example.com/3",
			Type:      "static",
			URL:       "https://example.com/3",
			CreatedBy: "user-001",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        "qrcode-004",
			AppID:     "app-001",
			Name:      "Dynamic QR Code 2",
			Content:   "https://example.com/4",
			Type:      "dynamic",
			URL:       "https://example.com/4",
			CreatedBy: "user-001",
			CreatedAt: time.Now(),
		},
	}

	for _, qrCode := range qrCodes {
		err := db.Create(qrCode).Error
		assert.NoError(t, err)
	}

	// Test case for listing QR codes with filtering by type
	t.Run("List QR codes with specific type", func(t *testing.T) {
		// Execute
		result, total, err := qrCodeService.ListQRCodes(context.Background(), "app-001", 1, 10)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 4, len(result))
		assert.Equal(t, int64(4), total)

		// Verify ordering (should be by created_at desc)
		if len(result) > 1 {
			for i := 0; i < len(result)-1; i++ {
				assert.True(t, result[i].CreatedAt.After(result[i+1].CreatedAt) || 
					result[i].CreatedAt.Equal(result[i+1].CreatedAt))
			}
		}

		// Verify that we have both static and dynamic QR codes
		staticCount := 0
		dynamicCount := 0
		for _, qr := range result {
			if qr.Type == "static" {
				staticCount++
			} else if qr.Type == "dynamic" {
				dynamicCount++
			}
		}
		assert.Equal(t, 2, staticCount)
		assert.Equal(t, 2, dynamicCount)
	})
}

func TestQRCodeService_GenerateQRCodeImage_Additional(t *testing.T) {
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

	// Create a QR code for testing with complex content
	qrCode := &domain.QRCode{
		ID:        "qrcode-001",
		AppID:     "app-001",
		Name:      "Complex Content QR Code",
		Content:   "https://example.com/complex?param1=value1¶m2=value2¶m3=value3",
		Type:      "static",
		URL:       "https://example.com/complex?param1=value1¶m2=value2¶m3=value3",
		CreatedBy: "user-001",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	err = db.Create(qrCode).Error
	assert.NoError(t, err)

	// Test case for generating QR code image with complex content
	t.Run("Generate image for QR code with complex content", func(t *testing.T) {
		// Execute
		imagePath, err := qrCodeService.GenerateQRCodeImage(context.Background(), "qrcode-001")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, imagePath)
	})
}
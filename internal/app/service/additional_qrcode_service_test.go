package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/shared/testutils"
)

// TestQRCodeServiceAdditional tests additional scenarios for the QRCodeService
func TestQRCodeServiceAdditional(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create QR code service with database connection
	qrCodeService := NewQRCodeService()

	// Replace the database connection with the test database
	qrCodeService.db = testDB

	// Test CreateQRCode with invalid type
	t.Run("CreateQRCodeInvalidType", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateQRCodeRequest{
			AppID:     "app_123",
			Name:      "Invalid Type QR Code",
			Content:   "https://example.com",
			Type:      "invalid",
			CreatedBy: "user_123",
		}

		qrCode, err := qrCodeService.CreateQRCode(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, qrCode)
		assert.Equal(t, "invalid QR code type", err.Error())
	})

	// Test CreateQRCode with dynamic type
	t.Run("CreateQRCodeDynamicType", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateQRCodeRequest{
			AppID:     "app_123",
			Name:      "Dynamic QR Code",
			Content:   "https://example.com",
			Type:      "dynamic",
			URL:       "https://example.com/redirect",
			CreatedBy: "user_123",
		}

		qrCode, err := qrCodeService.CreateQRCode(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, qrCode)
		assert.Equal(t, "dynamic", qrCode.Type)
		assert.Equal(t, "https://example.com/redirect", qrCode.URL)
	})

	// Test UpdateQRCode with non-existent ID
	t.Run("UpdateQRCodeNotFound", func(t *testing.T) {
		ctx := context.Background()
		req := &UpdateQRCodeRequest{
			Name: "Updated QR Code",
		}

		err := qrCodeService.UpdateQRCode(ctx, "non-existent-id", req)

		assert.Error(t, err)
		assert.Equal(t, "QR code not found", err.Error())
	})

	// Test DeleteQRCode with non-existent ID
	t.Run("DeleteQRCodeNotFound", func(t *testing.T) {
		ctx := context.Background()

		err := qrCodeService.DeleteQRCode(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Equal(t, "QR code not found", err.Error())
	})

	// Test GetQRCode with non-existent ID
	t.Run("GetQRCodeNotFound", func(t *testing.T) {
		ctx := context.Background()

		qrCode, err := qrCodeService.GetQRCode(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, qrCode)
		assert.Equal(t, "QR code not found", err.Error())
	})

	// Test ListQRCodes with invalid pagination
	t.Run("ListQRCodesInvalidPagination", func(t *testing.T) {
		ctx := context.Background()

		// Test with page = 0
		qrCodes, _, err := qrCodeService.ListQRCodes(ctx, "app_list", 0, 10)
		assert.NoError(t, err)
		assert.NotNil(t, qrCodes)
		// Just check it doesn't panic

		// Test with size = 0
		qrCodes, _, err = qrCodeService.ListQRCodes(ctx, "app_list", 1, 0)
		assert.NoError(t, err)
		assert.NotNil(t, qrCodes)
		// Default size should be 10

		// Test with size > 100
		qrCodes, _, err = qrCodeService.ListQRCodes(ctx, "app_list", 1, 150)
		assert.NoError(t, err)
		assert.NotNil(t, qrCodes)
		// Default size should be 10
	})

	// Test GenerateQRCodeImage with non-existent ID
	t.Run("GenerateQRCodeImageNotFound", func(t *testing.T) {
		ctx := context.Background()

		imagePath, err := qrCodeService.GenerateQRCodeImage(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Equal(t, "", imagePath)
		assert.Equal(t, "QR code not found", err.Error())
	})

	// Test UpdateQRCode with all fields
	t.Run("UpdateQRCodeAllFields", func(t *testing.T) {
		ctx := context.Background()

		// Create a QR code
		createReq := &CreateQRCodeRequest{
			AppID:     "app_123",
			Name:      "Update All Fields QR Code",
			Content:   "https://example.com",
			Type:      "static",
			CreatedBy: "user_123",
		}

		qrCode, err := qrCodeService.CreateQRCode(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, qrCode)

		// Update all fields
		updateReq := &UpdateQRCodeRequest{
			Name:    "Fully Updated QR Code",
			Content: "https://updated-example.com",
			URL:     "https://updated-example.com/redirect",
		}

		err = qrCodeService.UpdateQRCode(ctx, qrCode.ID, updateReq)
		assert.NoError(t, err)

		// Verify the update
		updatedQRCode, err := qrCodeService.GetQRCode(ctx, qrCode.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedQRCode)
		assert.Equal(t, "Fully Updated QR Code", updatedQRCode.Name)
		assert.Equal(t, "https://updated-example.com", updatedQRCode.Content)
		assert.Equal(t, "https://updated-example.com/redirect", updatedQRCode.URL)
	})
}
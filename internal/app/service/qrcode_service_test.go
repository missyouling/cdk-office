package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/shared/testutils"
)

// TestQRCodeService tests the QRCodeService
func TestQRCodeService(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create QR code service with database connection
	qrCodeService := NewQRCodeService()

	// Replace the database connection with the test database
	qrCodeService.db = testDB

	// Test CreateQRCode
	t.Run("CreateQRCode", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateQRCodeRequest{
			AppID:     "app_123",
			Name:      "Test QR Code",
			Content:   "https://example.com",
			Type:      "static",
			CreatedBy: "user_123",
		}

		qrCode, err := qrCodeService.CreateQRCode(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, qrCode)
		assert.Equal(t, "app_123", qrCode.AppID)
		assert.Equal(t, "Test QR Code", qrCode.Name)
		assert.Equal(t, "https://example.com", qrCode.Content)
		assert.Equal(t, "static", qrCode.Type)
	})

	// Test UpdateQRCode
	t.Run("UpdateQRCode", func(t *testing.T) {
		ctx := context.Background()

		// First create a QR code
		createReq := &CreateQRCodeRequest{
			AppID:     "app_123",
			Name:      "Update Test QR Code",
			Content:   "https://example.com",
			Type:      "static",
			CreatedBy: "user_123",
		}

		qrCode, err := qrCodeService.CreateQRCode(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, qrCode)

		// Now update the QR code
		updateReq := &UpdateQRCodeRequest{
			Name:    "Updated QR Code",
			Content: "https://updated-example.com",
		}

		err = qrCodeService.UpdateQRCode(ctx, qrCode.ID, updateReq)
		assert.NoError(t, err)
	})

	// Test DeleteQRCode
	t.Run("DeleteQRCode", func(t *testing.T) {
		ctx := context.Background()

		// First create a QR code
		createReq := &CreateQRCodeRequest{
			AppID:     "app_123",
			Name:      "Delete Test QR Code",
			Content:   "https://example.com",
			Type:      "static",
			CreatedBy: "user_123",
		}

		qrCode, err := qrCodeService.CreateQRCode(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, qrCode)

		// Now delete the QR code
		err = qrCodeService.DeleteQRCode(ctx, qrCode.ID)
		assert.NoError(t, err)
	})

	// Test ListQRCodes
	t.Run("ListQRCodes", func(t *testing.T) {
		ctx := context.Background()

		// Create a few QR codes
		for i := 1; i <= 3; i++ {
			req := &CreateQRCodeRequest{
				AppID:     "app_list",
				Name:      "List Test QR Code " + string(rune(i+'0')),
				Content:   "https://example.com/" + string(rune(i+'0')),
				Type:      "static",
				CreatedBy: "user_123",
			}

			_, err := qrCodeService.CreateQRCode(ctx, req)
			assert.NoError(t, err)
		}

		// List QR codes
		qrCodes, total, err := qrCodeService.ListQRCodes(ctx, "app_list", 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, qrCodes)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(qrCodes), 3)
	})

	// Test GetQRCode
	t.Run("GetQRCode", func(t *testing.T) {
		ctx := context.Background()

		// First create a QR code
		createReq := &CreateQRCodeRequest{
			AppID:     "app_123",
			Name:      "Get Test QR Code",
			Content:   "https://example.com",
			Type:      "static",
			CreatedBy: "user_123",
		}

		createdQRCode, err := qrCodeService.CreateQRCode(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, createdQRCode)

		// Now get the QR code
		retrievedQRCode, err := qrCodeService.GetQRCode(ctx, createdQRCode.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedQRCode)
		assert.Equal(t, createdQRCode.ID, retrievedQRCode.ID)
		assert.Equal(t, createdQRCode.Name, retrievedQRCode.Name)
	})
}
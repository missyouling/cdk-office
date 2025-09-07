package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cdk-office/internal/app/service"
)

func main() {
	// 创建批量二维码服务
	batchService := service.NewBatchQRCodeService()

	// 创建批量二维码请求
	req := &service.CreateBatchQRCodeRequest{
		AppID:       "app_test",
		Name:        "Test Batch",
		Description: "Test batch for QR code generation",
		Prefix:      "test",
		Count:       5,
		Type:        "static",
		URLTemplate: "https://example.com/test/{index}",
		CreatedBy:   "user_test",
	}

	// 创建批量二维码批次
	batch, err := batchService.CreateBatchQRCode(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to create batch QR code: %v", err)
	}

	fmt.Printf("Created batch QR code: %+v\n", batch)

	// 等待一段时间确保批次创建完成
	time.Sleep(1 * time.Second)

	// 生成批量二维码
	qrCodes, err := batchService.GenerateBatchQRCodes(context.Background(), batch.ID)
	if err != nil {
		log.Fatalf("Failed to generate batch QR codes: %v", err)
	}

	fmt.Printf("Generated %d QR codes\n", len(qrCodes))
	for i, qrCode := range qrCodes {
		fmt.Printf("QR Code %d: %+v\n", i+1, qrCode)
	}
}
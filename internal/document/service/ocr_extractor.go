package service

import (
	"cdk-office/internal/document/domain"
	"cdk-office/pkg/logger"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OCRExtractorInterface defines the interface for OCR extraction service
type OCRExtractorInterface interface {
	ExtractOCRContent(document *domain.Document) (string, error)
}

// OCRExtractor implements the OCRExtractorInterface
type OCRExtractor struct {
	storagePath string
}

// NewOCRExtractor creates a new instance of OCRExtractor
func NewOCRExtractor(storagePath string) *OCRExtractor {
	return &OCRExtractor{
		storagePath: storagePath,
	}
}

// ExtractOCRContent extracts content from a document using OCR
func (oe *OCRExtractor) ExtractOCRContent(document *domain.Document) (string, error) {
	// Determine file path
	filePath := document.FilePath
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(oe.storagePath, filePath)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Error("file not found", "file_path", filePath)
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	// Extract OCR content based on file type
	switch strings.ToLower(document.MimeType) {
	case "image/jpeg", "image/jpg", "image/png", "image/gif", "image/tiff", "image/bmp":
		return oe.extractImageOCRContent(filePath)
	case "application/pdf":
		return oe.extractPDFWithImageOCRContent(filePath)
	default:
		// OCR is not applicable for non-image files
		return "", fmt.Errorf("OCR not applicable for file type: %s", document.MimeType)
	}
}

// extractImageOCRContent extracts text from an image file using OCR
func (oe *OCRExtractor) extractImageOCRContent(filePath string) (string, error) {
	// Call the Python script to perform OCR using dots.ocr
	// The Python script is located in the same directory as this file.
	pythonScript := filepath.Join(filepath.Dir(filePath), "dots_ocr.py")
	
	// Check if the Python script exists
	if _, err := os.Stat(pythonScript); os.IsNotExist(err) {
		logger.Error("Python OCR script not found", "script", pythonScript)
		return "", fmt.Errorf("Python OCR script not found: %s", pythonScript)
	}
	
	// Execute the Python script
	cmd := exec.Command("python3", pythonScript, "image", filePath)
	output, err := cmd.Output()
	if err != nil {
		logger.Error("failed to execute Python OCR script", "error", err)
		return "", fmt.Errorf("failed to execute Python OCR script: %v", err)
	}
	
	// Parse the JSON output from the Python script
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		logger.Error("failed to parse OCR result", "error", err)
		return "", fmt.Errorf("failed to parse OCR result: %v", err)
	}
	
	// Check if there was an error in the OCR process
	if errorMsg, ok := result["error"]; ok {
		logger.Error("OCR process failed", "error", errorMsg)
		return "", fmt.Errorf("OCR process failed: %v", errorMsg)
	}
	
	// Extract the text from the result
	if text, ok := result["text"].(string); ok {
		return text, nil
	}
	
	logger.Error("OCR result does not contain text", "result", result)
	return "", fmt.Errorf("OCR result does not contain text")
}

// extractPDFWithImageOCRContent extracts text from a PDF file using OCR
func (oe *OCRExtractor) extractPDFWithImageOCRContent(filePath string) (string, error) {
	// Call the Python script to perform OCR using dots.ocr.
	// The Python script is located in the same directory as this file.
	pythonScript := filepath.Join(filepath.Dir(filePath), "dots_ocr.py")
	
	// Check if the Python script exists
	if _, err := os.Stat(pythonScript); os.IsNotExist(err) {
		logger.Error("Python OCR script not found", "script", pythonScript)
		return "", fmt.Errorf("Python OCR script not found: %s", pythonScript)
	}
	
	// Execute the Python script
	cmd := exec.Command("python3", pythonScript, "pdf", filePath)
	output, err := cmd.Output()
	if err != nil {
		logger.Error("failed to execute Python OCR script for PDF", "error", err)
		return "", fmt.Errorf("failed to execute Python OCR script for PDF: %v", err)
	}
	
	// Parse the JSON output from the Python script
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		logger.Error("failed to parse PDF OCR result", "error", err)
		return "", fmt.Errorf("failed to parse PDF OCR result: %v", err)
	}
	
	// Check if there was an error in the OCR process
	if errorMsg, ok := result["error"]; ok {
		logger.Error("PDF OCR process failed", "error", errorMsg)
		return "", fmt.Errorf("PDF OCR process failed: %v", errorMsg)
	}
	
	// Extract the text from the result
	// For PDFs, we concatenate text from all pages
	if pages, ok := result["pages"].([]interface{}); ok {
		var fullText string
		for _, page := range pages {
			if pageMap, ok := page.(map[string]interface{}); ok {
				if text, ok := pageMap["text"].(string); ok {
					fullText += text + "\n"
				}
			}
		}
		return fullText, nil
	}
	
	logger.Error("PDF OCR result does not contain pages", "result", result)
	return "", fmt.Errorf("PDF OCR result does not contain pages")
}
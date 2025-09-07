package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"cdk-office/pkg/logger"
)

// StorageServiceInterface defines the interface for storage service
type StorageServiceInterface interface {
	SaveFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID string) (string, error)
	DeleteFile(ctx context.Context, filePath string) error
	GetFile(ctx context.Context, filePath string) (io.ReadCloser, error)
}

// StorageService implements the StorageServiceInterface
type StorageService struct {
	storagePath string
}

// NewStorageService creates a new instance of StorageService
func NewStorageService() *StorageService {
	// In a real application, this path would be configurable
	storagePath := "/var/cdk-office/storage"
	
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		logger.Error("failed to create storage directory", "error", err)
		// In a real application, you might want to handle this error more gracefully
		// For now, we'll just log it and continue with a fallback path
		storagePath = "./storage"
		os.MkdirAll(storagePath, 0755)
	}
	
	return &StorageService{
		storagePath: storagePath,
	}
}

// SaveFile saves a file to the storage system
func (s *StorageService) SaveFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID string) (string, error) {
	// Generate a unique file name
	fileName := fmt.Sprintf("%s_%s%s", 
		userID, 
		time.Now().Format("20060102150405"), 
		filepath.Ext(header.Filename))
	
	// Create user directory if it doesn't exist
	userDir := filepath.Join(s.storagePath, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		logger.Error("failed to create user directory", "error", err)
		return "", errors.New("failed to save file")
	}
	
	// Create file path
	filePath := filepath.Join(userDir, fileName)
	
	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		logger.Error("failed to create file", "error", err)
		return "", errors.New("failed to save file")
	}
	defer dst.Close()
	
	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		logger.Error("failed to copy file content", "error", err)
		return "", errors.New("failed to save file")
	}
	
	return filePath, nil
}

// DeleteFile deletes a file from the storage system
func (s *StorageService) DeleteFile(ctx context.Context, filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("file not found")
	}
	
	// Delete the file
	if err := os.Remove(filePath); err != nil {
		logger.Error("failed to delete file", "error", err)
		return errors.New("failed to delete file")
	}
	
	return nil
}

// GetFile retrieves a file from the storage system
func (s *StorageService) GetFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.New("file not found")
	}
	
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("failed to open file", "error", err)
		return nil, errors.New("failed to get file")
	}
	
	return file, nil
}
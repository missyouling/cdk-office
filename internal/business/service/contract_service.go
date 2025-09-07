package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"cdk-office/internal/business/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/internal/shared/utils"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// ContractServiceInterface defines the interface for contract service
type ContractServiceInterface interface {
	CreateContract(ctx context.Context, req *CreateContractRequest) (*domain.Contract, error)
	UpdateContract(ctx context.Context, contractID string, req *UpdateContractRequest) error
	DeleteContract(ctx context.Context, contractID string) error
	ListContracts(ctx context.Context, teamID string, page, size int) ([]*domain.Contract, int64, error)
	GetContract(ctx context.Context, contractID string) (*domain.Contract, error)
	SignContract(ctx context.Context, contractID string, signerID string) error
}

// ContractService implements the ContractServiceInterface
type ContractService struct {
	db *gorm.DB
}

// NewContractService creates a new instance of ContractService
func NewContractService() *ContractService {
	return &ContractService{
		db: database.GetDB(),
	}
}

// CreateContractRequest represents the request for creating a contract
type CreateContractRequest struct {
	TeamID      string   `json:"team_id" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Content     string   `json:"content" binding:"required"`
	CreatedBy   string   `json:"created_by" binding:"required"`
	Signers     []string `json:"signers" binding:"required"`
}

// UpdateContractRequest represents the request for updating a contract
type UpdateContractRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// CreateContract creates a new contract
func (s *ContractService) CreateContract(ctx context.Context, req *CreateContractRequest) (*domain.Contract, error) {
	// Create new contract
	contract := &domain.Contract{
		ID:          utils.GenerateContractID(),
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		Status:      "draft",
		CreatedBy:   req.CreatedBy,
		TeamID:      req.TeamID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save contract to database
	if err := s.db.Create(contract).Error; err != nil {
		logger.Error("failed to create contract", "error", err)
		return nil, errors.New("failed to create contract")
	}

	return contract, nil
}

// UpdateContract updates an existing contract
func (s *ContractService) UpdateContract(ctx context.Context, contractID string, req *UpdateContractRequest) error {
	// Find contract by ID
	var contract domain.Contract
	if err := s.db.Where("id = ?", contractID).First(&contract).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("contract not found")
		}
		logger.Error("failed to find contract", "error", err)
		return errors.New("failed to update contract")
	}

	// Check if contract is in draft status
	if contract.Status != "draft" {
		return errors.New("only draft contracts can be updated")
	}

	// Update contract fields
	if req.Title != "" {
		contract.Title = req.Title
	}
	
	if req.Description != "" {
		contract.Description = req.Description
	}
	
	if req.Content != "" {
		contract.Content = req.Content
	}
	
	contract.UpdatedAt = time.Now()

	// Save updated contract to database
	if err := s.db.Save(&contract).Error; err != nil {
		logger.Error("failed to update contract", "error", err)
		return errors.New("failed to update contract")
	}

	return nil
}

// DeleteContract deletes a contract
func (s *ContractService) DeleteContract(ctx context.Context, contractID string) error {
	// Find contract by ID
	var contract domain.Contract
	if err := s.db.Where("id = ?", contractID).First(&contract).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("contract not found")
		}
		logger.Error("failed to find contract", "error", err)
		return errors.New("failed to delete contract")
	}

	// Check if contract is in draft status
	if contract.Status != "draft" {
		return errors.New("only draft contracts can be deleted")
	}

	// Delete contract from database
	if err := s.db.Delete(&contract).Error; err != nil {
		logger.Error("failed to delete contract", "error", err)
		return errors.New("failed to delete contract")
	}

	return nil
}

// ListContracts lists contracts with pagination
func (s *ContractService) ListContracts(ctx context.Context, teamID string, page, size int) ([]*domain.Contract, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build query
	dbQuery := s.db.Model(&domain.Contract{}).Where("team_id = ?", teamID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count contracts", "error", err)
		return nil, 0, errors.New("failed to list contracts")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var contracts []*domain.Contract
	if err := dbQuery.Find(&contracts).Error; err != nil {
		logger.Error("failed to list contracts", "error", err)
		return nil, 0, errors.New("failed to list contracts")
	}

	return contracts, total, nil
}

// GetContract retrieves a contract by ID
func (s *ContractService) GetContract(ctx context.Context, contractID string) (*domain.Contract, error) {
	var contract domain.Contract
	if err := s.db.Where("id = ?", contractID).First(&contract).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("contract not found")
		}
		logger.Error("failed to find contract", "error", err)
		return nil, errors.New("failed to get contract")
	}

	return &contract, nil
}

// SignContract signs a contract
func (s *ContractService) SignContract(ctx context.Context, contractID string, signerID string) error {
	// Find contract by ID
	var contract domain.Contract
	if err := s.db.Where("id = ?", contractID).First(&contract).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("contract not found")
		}
		logger.Error("failed to find contract", "error", err)
		return errors.New("failed to sign contract")
	}

	// Check if contract is in pending status
	if contract.Status != "pending" {
		return errors.New("contract is not in pending status")
	}

	// Check if signer is in the signers list
	signers, err := convertJSONToStringSlice(contract.Signers)
	if err != nil {
		logger.Error("failed to parse signers", "error", err)
		return errors.New("failed to sign contract")
	}

	isSigner := false
	for _, s := range signers {
		if s == signerID {
			isSigner = true
			break
		}
	}

	if !isSigner {
		return errors.New("user is not authorized to sign this contract")
	}

	// Check if signer has already signed
	signedBy, err := convertJSONToStringSlice(contract.SignedBy)
	if err != nil {
		logger.Error("failed to parse signed by", "error", err)
		return errors.New("failed to sign contract")
	}

	for _, s := range signedBy {
		if s == signerID {
			return errors.New("user has already signed this contract")
		}
	}

	// Add signer to signed by list
	signedBy = append(signedBy, signerID)
	contract.SignedBy = convertStringSliceToJSON(signedBy)

	// Check if all signers have signed
	if len(signedBy) == len(signers) {
		contract.Status = "completed"
	}

	contract.UpdatedAt = time.Now()

	// Save updated contract to database
	if err := s.db.Save(&contract).Error; err != nil {
		logger.Error("failed to update contract", "error", err)
		return errors.New("failed to sign contract")
	}

	return nil
}

// convertStringSliceToJSON converts a string slice to JSON string
func convertStringSliceToJSON(slice []string) string {
	// Convert string slice to JSON
	jsonData, err := json.Marshal(slice)
	if err != nil {
		logger.Error("failed to marshal string slice to JSON", "error", err)
		return "[]"
	}
	return string(jsonData)
}

// convertJSONToStringSlice converts JSON string to a string slice
func convertJSONToStringSlice(jsonStr string) ([]string, error) {
	// Convert JSON to string slice
	var slice []string
	if err := json.Unmarshal([]byte(jsonStr), &slice); err != nil {
		logger.Error("failed to unmarshal JSON to string slice", "error", err)
		return []string{}, err
	}
	return slice, nil
}
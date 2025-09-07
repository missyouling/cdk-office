package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/business/domain"
	"cdk-office/internal/business/service"
	"github.com/gin-gonic/gin"
)

// ContractHandlerInterface defines the interface for contract handler
type ContractHandlerInterface interface {
	CreateContract(c *gin.Context)
	UpdateContract(c *gin.Context)
	DeleteContract(c *gin.Context)
	ListContracts(c *gin.Context)
	GetContract(c *gin.Context)
	SignContract(c *gin.Context)
}

// ContractHandler implements the ContractHandlerInterface
type ContractHandler struct {
	contractService service.ContractServiceInterface
}

// NewContractHandler creates a new instance of ContractHandler
func NewContractHandler() *ContractHandler {
	return &ContractHandler{
		contractService: service.NewContractService(),
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

// SignContractRequest represents the request for signing a contract
type SignContractRequest struct {
	SignerID string `json:"signer_id" binding:"required"`
}

// ListContractsRequest represents the request for listing contracts
type ListContractsRequest struct {
	TeamID string `form:"team_id" binding:"required"`
	Page   int    `form:"page"`
	Size   int    `form:"size"`
}

// CreateContract handles creating a new contract
func (h *ContractHandler) CreateContract(c *gin.Context) {
	var req CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create contract
	contract, err := h.contractService.CreateContract(c.Request.Context(), &service.CreateContractRequest{
		TeamID:      req.TeamID,
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		CreatedBy:   req.CreatedBy,
		Signers:     req.Signers,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, contract)
}

// UpdateContract handles updating an existing contract
func (h *ContractHandler) UpdateContract(c *gin.Context) {
	contractID := c.Param("id")
	if contractID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contract id is required"})
		return
	}

	var req UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update contract
	if err := h.contractService.UpdateContract(c.Request.Context(), contractID, &service.UpdateContractRequest{
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
	}); err != nil {
		if err.Error() == "contract not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
			return
		}
		if err.Error() == "only draft contracts can be updated" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only draft contracts can be updated"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contract updated successfully"})
}

// DeleteContract handles deleting a contract
func (h *ContractHandler) DeleteContract(c *gin.Context) {
	contractID := c.Param("id")
	if contractID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contract id is required"})
		return
	}

	// Call service to delete contract
	if err := h.contractService.DeleteContract(c.Request.Context(), contractID); err != nil {
		if err.Error() == "contract not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
			return
		}
		if err.Error() == "only draft contracts can be deleted" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only draft contracts can be deleted"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contract deleted successfully"})
}

// ListContracts handles listing contracts with pagination
func (h *ContractHandler) ListContracts(c *gin.Context) {
	// Parse query parameters
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	page := 1
	size := 10

	var err error
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}

	if sizeStr != "" {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 1 || size > 100 {
			size = 10
		}
	}

	// Call service to list contracts
	contracts, total, err := h.contractService.ListContracts(c.Request.Context(), teamID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListContractsResponse{
		Items: contracts,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetContract handles retrieving a contract by ID
func (h *ContractHandler) GetContract(c *gin.Context) {
	contractID := c.Param("id")
	if contractID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contract id is required"})
		return
	}

	// Call service to get contract
	contract, err := h.contractService.GetContract(c.Request.Context(), contractID)
	if err != nil {
		if err.Error() == "contract not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, contract)
}

// SignContract handles signing a contract
func (h *ContractHandler) SignContract(c *gin.Context) {
	contractID := c.Param("id")
	if contractID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contract id is required"})
		return
	}

	var req SignContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to sign contract
	if err := h.contractService.SignContract(c.Request.Context(), contractID, req.SignerID); err != nil {
		if err.Error() == "contract not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
			return
		}
		if err.Error() == "contract is not in pending status" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "contract is not in pending status"})
			return
		}
		if err.Error() == "user is not authorized to sign this contract" {
			c.JSON(http.StatusForbidden, gin.H{"error": "user is not authorized to sign this contract"})
			return
		}
		if err.Error() == "user has already signed this contract" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user has already signed this contract"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contract signed successfully"})
}

// ListContractsResponse represents the response for listing contracts
type ListContractsResponse struct {
	Items []*domain.Contract `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}
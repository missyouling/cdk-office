package formdesigner_test

import (
	"context"
	"testing"
	"time"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

func TestFormDesignerService_CreateFormDesign(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormDesign{}, &domain.Application{})

	formDesignerService := service.NewFormDesignerService()

	// Create a test application first
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "form",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		request       *service.CreateFormDesignRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid form design creation",
			request: &service.CreateFormDesignRequest{
				AppID:       "app-001",
				Name:        "Test Form Design",
				Description: "Test form design description",
				Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
				Config:      `{"theme": "default"}`,
				CreatedBy:   "user-001",
			},
			expectError: false,
		},
		{
			name: "Non-existent application",
			request: &service.CreateFormDesignRequest{
				AppID:       "non-existent-app",
				Name:        "Orphan Form Design",
				Description: "Orphan form design description",
				Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
				Config:      `{"theme": "default"}`,
				CreatedBy:   "user-001",
			},
			expectError: false, // Note: This doesn't validate app existence in the service
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			form, err := formDesignerService.CreateFormDesign(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, form)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, form)
				assert.NotEmpty(t, form.ID)
				assert.Equal(t, tt.request.AppID, form.AppID)
				assert.Equal(t, tt.request.Name, form.Name)
				assert.Equal(t, tt.request.Description, form.Description)
				assert.Equal(t, tt.request.Schema, form.Schema)
				assert.Equal(t, tt.request.Config, form.Config)
				assert.True(t, form.IsActive)
				assert.False(t, form.IsPublished)
				assert.Equal(t, tt.request.CreatedBy, form.CreatedBy)
				assert.WithinDuration(t, time.Now(), form.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), form.UpdatedAt, time.Second)
			}
		})
	}
}

func TestFormDesignerService_UpdateFormDesign(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormDesign{}, &domain.Application{})

	formDesignerService := service.NewFormDesignerService()

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "form",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a form design for testing
	form := &domain.FormDesign{
		ID:          "form-001",
		AppID:       "app-001",
		Name:        "Original Form Design",
		Description: "Original description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"theme": "default"}`,
		IsActive:    true,
		IsPublished: false,
		CreatedBy:   "user-001",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	err = db.Table("form_designs").Create(form).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		formID        string
		request       *service.UpdateFormDesignRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name:   "Valid form design update",
			formID: "form-001",
			request: &service.UpdateFormDesignRequest{
				Name:        "Updated Form Design",
				Description: "Updated description",
				Schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}}}`,
				Config:      `{"theme": "dark"}`,
				IsActive:    boolPtr(false),
			},
			expectError: false,
		},
		{
			name:        "Update non-existent form design",
			formID:      "non-existent-id",
			request:     &service.UpdateFormDesignRequest{Name: "Updated Name"},
			expectError: true,
			errorMessage: "form design not found",
		},
		{
			name:   "Update published form design",
			formID: "form-002",
			request: &service.UpdateFormDesignRequest{
				Name: "Updated Published Form",
			},
			expectError: true,
			errorMessage: "cannot update published form design",
		},
		{
			name:   "Partial update - name only",
			formID: "form-001",
			request: &service.UpdateFormDesignRequest{
				Name: "Name Only Update",
			},
			expectError: false,
		},
	}

	// Create a published form design for testing
	publishedForm := &domain.FormDesign{
		ID:          "form-002",
		AppID:       "app-001",
		Name:        "Published Form Design",
		Description: "Published description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"theme": "default"}`,
		IsActive:    true,
		IsPublished: true,
		CreatedBy:   "user-001",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	err = db.Table("form_designs").Create(publishedForm).Error
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := formDesignerService.UpdateFormDesign(context.Background(), tt.formID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify the update
				updatedForm, getErr := formDesignerService.GetFormDesign(context.Background(), tt.formID)
				assert.NoError(t, getErr)
				assert.NotNil(t, updatedForm)

				// Check updated fields
				if tt.request.Name != "" {
					assert.Equal(t, tt.request.Name, updatedForm.Name)
				}
				if tt.request.Description != "" {
					assert.Equal(t, tt.request.Description, updatedForm.Description)
				}
				if tt.request.Schema != "" {
					assert.Equal(t, tt.request.Schema, updatedForm.Schema)
				}
				if tt.request.Config != "" {
					assert.Equal(t, tt.request.Config, updatedForm.Config)
				}
				if tt.request.IsActive != nil {
					assert.Equal(t, *tt.request.IsActive, updatedForm.IsActive)
				}
				// UpdatedAt should be changed
				assert.True(t, updatedForm.UpdatedAt.After(form.UpdatedAt))
			}
		})
	}
}

func TestFormDesignerService_DeleteFormDesign(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormDesign{}, &domain.Application{})

	formDesignerService := service.NewFormDesignerService()

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "form",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a form design for testing
	form := &domain.FormDesign{
		ID:          "form-001",
		AppID:       "app-001",
		Name:        "Test Form Design",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"theme": "default"}`,
		IsActive:    true,
		IsPublished: false,
		CreatedBy:   "user-001",
	}
	err = db.Table("form_designs").Create(form).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		formID        string
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "Delete non-existent form design",
			formID:      "non-existent-id",
			expectError: true,
			errorMessage: "form design not found",
		},
		{
			name:        "Delete published form design",
			formID:      "form-002",
			expectError: true,
			errorMessage: "cannot delete published form design",
		},
		{
			name:        "Valid form design deletion",
			formID:      "form-001",
			expectError: false,
		},
	}

	// Create a published form design for testing
	publishedForm := &domain.FormDesign{
		ID:          "form-002",
		AppID:       "app-001",
		Name:        "Published Form Design",
		Description: "Published description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"theme": "default"}`,
		IsActive:    true,
		IsPublished: true,
		CreatedBy:   "user-001",
	}
	err = db.Table("form_designs").Create(publishedForm).Error
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := formDesignerService.DeleteFormDesign(context.Background(), tt.formID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify form design is deleted
				_, getErr := formDesignerService.GetFormDesign(context.Background(), tt.formID)
				assert.Error(t, getErr)
				assert.Equal(t, "form design not found", getErr.Error())
			}
		})
	}
}

func TestFormDesignerService_ListFormDesigns(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormDesign{}, &domain.Application{})

	formDesignerService := service.NewFormDesignerService()

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "form",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create test form designs
	forms := []*domain.FormDesign{
		{
			ID:          "form-001",
			AppID:       "app-001",
			Name:        "Form Design 1",
			Description: "First form design",
			Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
			Config:      `{"theme": "default"}`,
			IsActive:    true,
			IsPublished: false,
			CreatedBy:   "user-001",
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          "form-002",
			AppID:       "app-001",
			Name:        "Form Design 2",
			Description: "Second form design",
			Schema:      `{"type": "object", "properties": {"age": {"type": "number"}}}`,
			Config:      `{"theme": "dark"}`,
			IsActive:    true,
			IsPublished: true,
			CreatedBy:   "user-001",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "form-003",
			AppID:       "app-001",
			Name:        "Form Design 3",
			Description: "Third form design",
			Schema:      `{"type": "object", "properties": {"email": {"type": "string"}}}`,
			Config:      `{"theme": "light"}`,
			IsActive:    false,
			IsPublished: false,
			CreatedBy:   "user-001",
			CreatedAt:   time.Now(),
		},
	}

	for _, form := range forms {
		err := db.Table("form_designs").Create(form).Error
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
			result, total, err := formDesignerService.ListFormDesigns(context.Background(), tt.appID, tt.page, tt.size)

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

func TestFormDesignerService_GetFormDesign(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormDesign{}, &domain.Application{})

	formDesignerService := service.NewFormDesignerService()

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "form",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a form design for testing
	form := &domain.FormDesign{
		ID:          "form-001",
		AppID:       "app-001",
		Name:        "Test Form Design",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"theme": "default"}`,
		IsActive:    true,
		IsPublished: false,
		CreatedBy:   "user-001",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	err = db.Table("form_designs").Create(form).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		formID        string
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "Get non-existent form design",
			formID:      "non-existent-id",
			expectError: true,
			errorMessage: "form design not found",
		},
		{
			name:        "Get existing form design",
			formID:      "form-001",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := formDesignerService.GetFormDesign(context.Background(), tt.formID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, form.ID, result.ID)
				assert.Equal(t, form.AppID, result.AppID)
				assert.Equal(t, form.Name, result.Name)
				assert.Equal(t, form.Description, result.Description)
				assert.Equal(t, form.Schema, result.Schema)
				assert.Equal(t, form.Config, result.Config)
				assert.Equal(t, form.IsActive, result.IsActive)
				assert.Equal(t, form.IsPublished, result.IsPublished)
				assert.Equal(t, form.CreatedBy, result.CreatedBy)
				assert.Equal(t, form.CreatedAt.Unix(), result.CreatedAt.Unix())
				assert.Equal(t, form.UpdatedAt.Unix(), result.UpdatedAt.Unix())
			}
		})
	}
}

func TestFormDesignerService_PublishFormDesign(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormDesign{}, &domain.Application{})

	formDesignerService := service.NewFormDesignerService()

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "form",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a form design for testing
	form := &domain.FormDesign{
		ID:          "form-001",
		AppID:       "app-001",
		Name:        "Test Form Design",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"theme": "default"}`,
		IsActive:    true,
		IsPublished: false,
		CreatedBy:   "user-001",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	err = db.Table("form_designs").Create(form).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		formID        string
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "Publish non-existent form design",
			formID:      "non-existent-id",
			expectError: true,
			errorMessage: "form design not found",
		},
		{
			name:        "Publish already published form design",
			formID:      "form-002",
			expectError: true,
			errorMessage: "form design is already published",
		},
		{
			name:        "Valid form design publishing",
			formID:      "form-001",
			expectError: false,
		},
	}

	// Create an already published form design for testing
	publishedForm := &domain.FormDesign{
		ID:          "form-002",
		AppID:       "app-001",
		Name:        "Published Form Design",
		Description: "Published description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"theme": "default"}`,
		IsActive:    true,
		IsPublished: true,
		CreatedBy:   "user-001",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	err = db.Table("form_designs").Create(publishedForm).Error
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := formDesignerService.PublishFormDesign(context.Background(), tt.formID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify the form design is published
				publishedForm, getErr := formDesignerService.GetFormDesign(context.Background(), tt.formID)
				assert.NoError(t, getErr)
				assert.NotNil(t, publishedForm)
				assert.True(t, publishedForm.IsPublished)
				// UpdatedAt should be changed
				assert.True(t, publishedForm.UpdatedAt.After(form.UpdatedAt))
			}
		})
	}
}

// Helper function to create a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}
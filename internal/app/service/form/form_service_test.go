package form_test

import (
	"context"
	"testing"
	"time"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

func TestFormService_CreateForm(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormData{}, &domain.Application{})

	formService := service.NewFormService()

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
		request       *service.CreateFormRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid form creation",
			request: &service.CreateFormRequest{
				AppID:       "app-001",
				Name:        "Test Form",
				Description: "Test form description",
				Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
				CreatedBy:   "user-001",
			},
			expectError: false,
		},
		{
			name: "Non-existent application",
			request: &service.CreateFormRequest{
				AppID:       "non-existent-app",
				Name:        "Orphan Form",
				Description: "Orphan form description",
				Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
				CreatedBy:   "user-001",
			},
			expectError: false, // Note: This doesn't validate app existence in the service
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			form, err := formService.CreateForm(context.Background(), tt.request)

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
				assert.True(t, form.IsActive)
				assert.Equal(t, tt.request.CreatedBy, form.CreatedBy)
				assert.WithinDuration(t, time.Now(), form.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), form.UpdatedAt, time.Second)
			}
		})
	}
}

func TestFormService_UpdateForm(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormData{}, &domain.Application{})

	formService := service.NewFormService()

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

	// Create a form for testing
	form := &domain.FormData{
		ID:          "form-001",
		AppID:       "app-001",
		Name:        "Original Form",
		Description: "Original description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		IsActive:    true,
		CreatedBy:   "user-001",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	err = db.Create(form).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		formID        string
		request       *service.UpdateFormRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name:   "Valid form update",
			formID: "form-001",
			request: &service.UpdateFormRequest{
				Name:        "Updated Form",
				Description: "Updated description",
				Schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}}}`,
				IsActive:    boolPtr(false),
			},
			expectError: false,
		},
		{
			name:        "Update non-existent form",
			formID:      "non-existent-id",
			request:     &service.UpdateFormRequest{Name: "Updated Name"},
			expectError: true,
			errorMessage: "form not found",
		},
		{
			name:   "Partial update - name only",
			formID: "form-001",
			request: &service.UpdateFormRequest{
				Name: "Name Only Update",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := formService.UpdateForm(context.Background(), tt.formID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify the update
				updatedForm, getErr := formService.GetForm(context.Background(), tt.formID)
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
				if tt.request.IsActive != nil {
					assert.Equal(t, *tt.request.IsActive, updatedForm.IsActive)
				}
				// UpdatedAt should be changed
				assert.True(t, updatedForm.UpdatedAt.After(form.UpdatedAt))
			}
		})
	}
}

func TestFormService_DeleteForm(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormData{}, &domain.Application{})

	formService := service.NewFormService()

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

	// Create a form for testing
	form := &domain.FormData{
		ID:          "form-001",
		AppID:       "app-001",
		Name:        "Test Form",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		IsActive:    true,
		CreatedBy:   "user-001",
	}
	err = db.Create(form).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		formID        string
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "Delete non-existent form",
			formID:      "non-existent-id",
			expectError: true,
			errorMessage: "form not found",
		},
		{
			name:        "Valid form deletion",
			formID:      "form-001",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := formService.DeleteForm(context.Background(), tt.formID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify form is deleted
				_, getErr := formService.GetForm(context.Background(), tt.formID)
				assert.Error(t, getErr)
				assert.Equal(t, "form not found", getErr.Error())
			}
		})
	}
}

func TestFormService_ListForms(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormData{}, &domain.Application{})

	formService := service.NewFormService()

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

	// Create test forms
	forms := []*domain.FormData{
		{
			ID:          "form-001",
			AppID:       "app-001",
			Name:        "Form 1",
			Description: "First form",
			Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
			IsActive:    true,
			CreatedBy:   "user-001",
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          "form-002",
			AppID:       "app-001",
			Name:        "Form 2",
			Description: "Second form",
			Schema:      `{"type": "object", "properties": {"age": {"type": "number"}}}`,
			IsActive:    true,
			CreatedBy:   "user-001",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "form-003",
			AppID:       "app-001",
			Name:        "Form 3",
			Description: "Third form",
			Schema:      `{"type": "object", "properties": {"email": {"type": "string"}}}`,
			IsActive:    false,
			CreatedBy:   "user-001",
			CreatedAt:   time.Now(),
		},
	}

	for _, form := range forms {
		err := db.Create(form).Error
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
			result, total, err := formService.ListForms(context.Background(), tt.appID, tt.page, tt.size)

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

func TestFormService_GetForm(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormData{}, &domain.Application{})

	formService := service.NewFormService()

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

	// Create a form for testing
	form := &domain.FormData{
		ID:          "form-001",
		AppID:       "app-001",
		Name:        "Test Form",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		IsActive:    true,
		CreatedBy:   "user-001",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	err = db.Create(form).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		formID        string
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "Get non-existent form",
			formID:      "non-existent-id",
			expectError: true,
			errorMessage: "form not found",
		},
		{
			name:        "Get existing form",
			formID:      "form-001",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := formService.GetForm(context.Background(), tt.formID)

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
				assert.Equal(t, form.IsActive, result.IsActive)
				assert.Equal(t, form.CreatedBy, result.CreatedBy)
				assert.Equal(t, form.CreatedAt.Unix(), result.CreatedAt.Unix())
				assert.Equal(t, form.UpdatedAt.Unix(), result.UpdatedAt.Unix())
			}
		})
	}
}

func TestFormService_SubmitFormData(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormData{}, &domain.FormDataEntry{}, &domain.Application{})

	formService := service.NewFormService()

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

	// Create a form for testing
	form := &domain.FormData{
		ID:          "form-001",
		AppID:       "app-001",
		Name:        "Test Form",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		IsActive:    true,
		CreatedBy:   "user-001",
	}
	err = db.Create(form).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		request       *service.SubmitFormDataRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid form data submission",
			request: &service.SubmitFormDataRequest{
				FormID:    "form-001",
				Data:      `{"name": "John Doe"}`,
				CreatedBy: "user-001",
			},
			expectError: false,
		},
		{
			name: "Submit to non-existent form",
			request: &service.SubmitFormDataRequest{
				FormID:    "non-existent-id",
				Data:      `{"name": "Jane Doe"}`,
				CreatedBy: "user-001",
			},
			expectError: true,
			errorMessage: "form not found or inactive",
		},
		{
			name: "Submit to inactive form",
			request: &service.SubmitFormDataRequest{
				FormID:    "form-002",
				Data:      `{"name": "Jane Doe"}`,
				CreatedBy: "user-001",
			},
			expectError: true,
			errorMessage: "form not found or inactive",
		},
	}

	// Create an inactive form for testing
	inactiveForm := &domain.FormData{
		ID:          "form-002",
		AppID:       "app-001",
		Name:        "Inactive Form",
		Description: "Inactive form description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		IsActive:    false,
		CreatedBy:   "user-001",
	}
	err = db.Create(inactiveForm).Error
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			entry, err := formService.SubmitFormData(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, entry)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entry)
				assert.NotEmpty(t, entry.ID)
				assert.Equal(t, tt.request.FormID, entry.FormID)
				assert.Equal(t, tt.request.Data, entry.Data)
				assert.Equal(t, tt.request.CreatedBy, entry.CreatedBy)
				assert.WithinDuration(t, time.Now(), entry.CreatedAt, time.Second)
			}
		})
	}
}

func TestFormService_ListFormDataEntries(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.FormData{}, &domain.FormDataEntry{}, &domain.Application{})

	formService := service.NewFormService()

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

	// Create a form for testing
	form := &domain.FormData{
		ID:          "form-001",
		AppID:       "app-001",
		Name:        "Test Form",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		IsActive:    true,
		CreatedBy:   "user-001",
	}
	err = db.Create(form).Error
	assert.NoError(t, err)

	// Create test form data entries
	entries := []*domain.FormDataEntry{
		{
			ID:        "entry-001",
			FormID:    "form-001",
			Data:      `{"name": "John Doe"}`,
			CreatedBy: "user-001",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "entry-002",
			FormID:    "form-001",
			Data:      `{"name": "Jane Smith"}`,
			CreatedBy: "user-002",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        "entry-003",
			FormID:    "form-001",
			Data:      `{"name": "Bob Johnson"}`,
			CreatedBy: "user-003",
			CreatedAt: time.Now(),
		},
	}

	for _, entry := range entries {
		err := db.Create(entry).Error
		assert.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name              string
		formID            string
		page              int
		size              int
		expectedCount     int
		totalCount        int64
		expectError       bool
		errorMessage      string
	}{
		{
			name:          "List first page",
			formID:        "form-001",
			page:          1,
			size:          2,
			expectedCount: 2,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List second page",
			formID:        "form-001",
			page:          2,
			size:          2,
			expectedCount: 1,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List with large page size",
			formID:        "form-001",
			page:          1,
			size:          10,
			expectedCount: 3,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List with zero page",
			formID:        "form-001",
			page:          0,
			size:          10,
			expectedCount: 3,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List non-existent form",
			formID:        "non-existent-id",
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
			result, total, err := formService.ListFormDataEntries(context.Background(), tt.formID, tt.page, tt.size)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
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

// Helper function to create a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}
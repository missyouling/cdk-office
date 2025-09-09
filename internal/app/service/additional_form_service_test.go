package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/shared/testutils"
)

// TestFormServiceAdditional tests additional scenarios for the FormService
func TestFormServiceAdditional(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create form service with database connection
	formService := NewFormService()

	// Replace the database connection with the test database
	formService.db = testDB

	// Test UpdateForm with non-existent ID
	t.Run("UpdateFormNotFound", func(t *testing.T) {
		ctx := context.Background()
		isActive := true
		req := &UpdateFormRequest{
			Name:     "Updated Form",
			IsActive: &isActive,
		}

		err := formService.UpdateForm(ctx, "non-existent-id", req)

		assert.Error(t, err)
		assert.Equal(t, "form not found", err.Error())
	})

	// Test DeleteForm with non-existent ID
	t.Run("DeleteFormNotFound", func(t *testing.T) {
		ctx := context.Background()

		err := formService.DeleteForm(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Equal(t, "form not found", err.Error())
	})

	// Test GetForm with non-existent ID
	t.Run("GetFormNotFound", func(t *testing.T) {
		ctx := context.Background()

		form, err := formService.GetForm(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, form)
		assert.Equal(t, "form not found", err.Error())
	})

	// Test ListForms with invalid pagination
	t.Run("ListFormsInvalidPagination", func(t *testing.T) {
		ctx := context.Background()

		// Test with page = 0
		forms, _, err := formService.ListForms(ctx, "app_list", 0, 10)
		assert.NoError(t, err)
		assert.NotNil(t, forms)
		// Just check it doesn't panic

		// Test with size = 0
		forms, _, err = formService.ListForms(ctx, "app_list", 1, 0)
		assert.NoError(t, err)
		assert.NotNil(t, forms)
		// Default size should be 10

		// Test with size > 100
		forms, _, err = formService.ListForms(ctx, "app_list", 1, 150)
		assert.NoError(t, err)
		assert.NotNil(t, forms)
		// Default size should be 10
	})

	// Test SubmitFormData with non-existent form
	t.Run("SubmitFormDataFormNotFound", func(t *testing.T) {
		ctx := context.Background()
		req := &SubmitFormDataRequest{
			FormID:    "non-existent-id",
			Data:      "{\"field1\":\"test value\"}",
			CreatedBy: "user_123",
		}

		entry, err := formService.SubmitFormData(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, entry)
		assert.Equal(t, "form not found or inactive", err.Error())
	})

	// Test SubmitFormData with inactive form
	t.Run("SubmitFormDataFormInactive", func(t *testing.T) {
		ctx := context.Background()

		// First create a form
		createReq := &CreateFormRequest{
			AppID:     "app_123",
			Name:      "Inactive Test Form",
			Schema:    "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}",
			CreatedBy: "user_123",
		}

		form, err := formService.CreateForm(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, form)

		// Deactivate the form
		isActive := false
		updateReq := &UpdateFormRequest{
			IsActive: &isActive,
		}

		err = formService.UpdateForm(ctx, form.ID, updateReq)
		assert.NoError(t, err)

		// Try to submit data to inactive form
		submitReq := &SubmitFormDataRequest{
			FormID:    form.ID,
			Data:      "{\"field1\":\"test value\"}",
			CreatedBy: "user_123",
		}

		entry, err := formService.SubmitFormData(ctx, submitReq)

		assert.Error(t, err)
		assert.Nil(t, entry)
		assert.Equal(t, "form not found or inactive", err.Error())
	})

	// Test ListFormDataEntries with non-existent form
	t.Run("ListFormDataEntriesFormNotFound", func(t *testing.T) {
		ctx := context.Background()

		entries, total, err := formService.ListFormDataEntries(ctx, "non-existent-id", 1, 10)

		assert.NoError(t, err) // Should not error, just return empty list
		assert.NotNil(t, entries)
		assert.Equal(t, int64(0), total)
		assert.Len(t, entries, 0)
	})

	// Test ListFormDataEntries with invalid pagination
	t.Run("ListFormDataEntriesInvalidPagination", func(t *testing.T) {
		ctx := context.Background()

		// First create a form
		createReq := &CreateFormRequest{
			AppID:     "app_123",
			Name:      "Pagination Test Form",
			Schema:    "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}",
			CreatedBy: "user_123",
		}

		form, err := formService.CreateForm(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, form)

		// Test with page = 0
		entries, _, err := formService.ListFormDataEntries(ctx, form.ID, 0, 10)
		assert.NoError(t, err)
		assert.NotNil(t, entries)
		// Just check it doesn't panic

		// Test with size = 0
		entries, _, err = formService.ListFormDataEntries(ctx, form.ID, 1, 0)
		assert.NoError(t, err)
		assert.NotNil(t, entries)
		// Default size should be 10

		// Test with size > 100
		entries, _, err = formService.ListFormDataEntries(ctx, form.ID, 1, 150)
		assert.NoError(t, err)
		assert.NotNil(t, entries)
		// Default size should be 10
	})

	// Test UpdateForm with all fields
	t.Run("UpdateFormAllFields", func(t *testing.T) {
		ctx := context.Background()

		// Create a form
		createReq := &CreateFormRequest{
			AppID:     "app_123",
			Name:      "Update All Fields Form",
			Schema:    "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}",
			CreatedBy: "user_123",
		}

		form, err := formService.CreateForm(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, form)

		// Update all fields
		isActive := false
		updateReq := &UpdateFormRequest{
			Name:        "Fully Updated Form",
			Description: "Updated description",
			Schema:      "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"},{\"name\":\"field2\",\"type\":\"number\"}]}",
			IsActive:    &isActive,
		}

		err = formService.UpdateForm(ctx, form.ID, updateReq)
		assert.NoError(t, err)

		// Verify the update
		updatedForm, err := formService.GetForm(ctx, form.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedForm)
		assert.Equal(t, "Fully Updated Form", updatedForm.Name)
		assert.Equal(t, "Updated description", updatedForm.Description)
		assert.Equal(t, "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"},{\"name\":\"field2\",\"type\":\"number\"}]}", updatedForm.Schema)
		assert.Equal(t, false, updatedForm.IsActive)
	})

	// Test multiple form submissions
	t.Run("MultipleFormSubmissions", func(t *testing.T) {
		ctx := context.Background()

		// Create a form
		createReq := &CreateFormRequest{
			AppID:     "app_123",
			Name:      "Multiple Submissions Test Form",
			Schema:    "{\"fields\":[{\"name\":\"name\",\"type\":\"text\"},{\"name\":\"email\",\"type\":\"email\"}]}",
			CreatedBy: "user_123",
		}

		form, err := formService.CreateForm(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, form)

		// Submit multiple form data entries
		testData := []struct {
			name  string
			email string
		}{
			{"John Doe", "john@example.com"},
			{"Jane Smith", "jane@example.com"},
			{"Bob Johnson", "bob@example.com"},
		}

		for _, data := range testData {
			// Add a small delay to ensure unique IDs
			time.Sleep(1 * time.Second)
			
			submitReq := &SubmitFormDataRequest{
				FormID:    form.ID,
				Data:      "{\"name\":\"" + data.name + "\",\"email\":\"" + data.email + "\"}",
				CreatedBy: "user_123",
			}

			entry, err := formService.SubmitFormData(ctx, submitReq)
			assert.NoError(t, err)
			assert.NotNil(t, entry)
		}

		// List form data entries
		entries, total, err := formService.ListFormDataEntries(ctx, form.ID, 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, entries)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(entries), 3)
	})
}
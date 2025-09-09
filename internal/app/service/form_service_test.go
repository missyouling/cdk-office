package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/shared/testutils"
)

// TestFormService tests the FormService
func TestFormService(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create form service with database connection
	formService := NewFormService()

	// Replace the database connection with the test database
	formService.db = testDB

	// Test CreateForm
	t.Run("CreateForm", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateFormRequest{
			AppID:       "app_123",
			Name:        "Test Form",
			Description: "Test form description",
			Schema:      "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}",
			CreatedBy:   "user_123",
		}

		form, err := formService.CreateForm(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, form)
		assert.Equal(t, "app_123", form.AppID)
		assert.Equal(t, "Test Form", form.Name)
		assert.Equal(t, "Test form description", form.Description)
		assert.Equal(t, "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}", form.Schema)
		assert.True(t, form.IsActive)
	})

	// Test UpdateForm
	t.Run("UpdateForm", func(t *testing.T) {
		ctx := context.Background()

		// First create a form
		createReq := &CreateFormRequest{
			AppID:     "app_123",
			Name:      "Update Test Form",
			Schema:    "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}",
			CreatedBy: "user_123",
		}

		form, err := formService.CreateForm(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, form)

		// Now update the form
		isActive := false
		updateReq := &UpdateFormRequest{
			Name:        "Updated Form",
			Description: "Updated description",
			Schema:      "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"},{\"name\":\"field2\",\"type\":\"number\"}]}",
			IsActive:    &isActive,
		}

		err = formService.UpdateForm(ctx, form.ID, updateReq)
		assert.NoError(t, err)
	})

	// Test DeleteForm
	t.Run("DeleteForm", func(t *testing.T) {
		ctx := context.Background()

		// First create a form
		createReq := &CreateFormRequest{
			AppID:     "app_123",
			Name:      "Delete Test Form",
			Schema:    "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}",
			CreatedBy: "user_123",
		}

		form, err := formService.CreateForm(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, form)

		// Now delete the form
		err = formService.DeleteForm(ctx, form.ID)
		assert.NoError(t, err)
	})

	// Test ListForms
	t.Run("ListForms", func(t *testing.T) {
		ctx := context.Background()

		// Create a few forms
		for i := 1; i <= 3; i++ {
			req := &CreateFormRequest{
				AppID:     "app_list",
				Name:      "List Test Form " + string(rune(i+'0')),
				Schema:    "{\"fields\":[{\"name\":\"field" + string(rune(i+'0')) + "\",\"type\":\"text\"}]}",
				CreatedBy: "user_123",
			}

			_, err := formService.CreateForm(ctx, req)
			assert.NoError(t, err)
		}

		// List forms
		forms, total, err := formService.ListForms(ctx, "app_list", 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, forms)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(forms), 3)
	})

	// Test GetForm
	t.Run("GetForm", func(t *testing.T) {
		ctx := context.Background()

		// First create a form
		createReq := &CreateFormRequest{
			AppID:     "app_123",
			Name:      "Get Test Form",
			Schema:    "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}",
			CreatedBy: "user_123",
		}

		createdForm, err := formService.CreateForm(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, createdForm)

		// Now get the form
		retrievedForm, err := formService.GetForm(ctx, createdForm.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedForm)
		assert.Equal(t, createdForm.ID, retrievedForm.ID)
		assert.Equal(t, createdForm.Name, retrievedForm.Name)
	})

	// Test SubmitFormData
	t.Run("SubmitFormData", func(t *testing.T) {
		ctx := context.Background()

		// First create a form
		createReq := &CreateFormRequest{
			AppID:     "app_123",
			Name:      "Submit Test Form",
			Schema:    "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}",
			CreatedBy: "user_123",
		}

		form, err := formService.CreateForm(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, form)

		// Now submit form data
		submitReq := &SubmitFormDataRequest{
			FormID:    form.ID,
			Data:      "{\"field1\":\"test value\"}",
			CreatedBy: "user_123",
		}

		entry, err := formService.SubmitFormData(ctx, submitReq)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, form.ID, entry.FormID)
		assert.Equal(t, "{\"field1\":\"test value\"}", entry.Data)
	})

	// Test ListFormDataEntries
	t.Run("ListFormDataEntries", func(t *testing.T) {
		ctx := context.Background()

		// First create a form
		createReq := &CreateFormRequest{
			AppID:     "app_123",
			Name:      "List Entries Test Form",
			Schema:    "{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}",
			CreatedBy: "user_123",
		}

		form, err := formService.CreateForm(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, form)

		// Submit a few form data entries
		for i := 1; i <= 3; i++ {
			// Add a small delay to ensure unique IDs
			time.Sleep(1 * time.Second)
			
			submitReq := &SubmitFormDataRequest{
				FormID:    form.ID,
				Data:      "{\"field1\":\"test value " + string(rune(i+'0')) + "\"}",
				CreatedBy: "user_123",
			}

			_, err := formService.SubmitFormData(ctx, submitReq)
			assert.NoError(t, err)
		}

		// List form data entries
		entries, total, err := formService.ListFormDataEntries(ctx, form.ID, 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, entries)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(entries), 3)
	})
}
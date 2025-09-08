package datacollection_test

import (
	"context"
	"testing"
	"time"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDataCollectionService_CreateDataCollection(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.DataCollection{}, &domain.DataCollectionEntry{}, &domain.Application{})

	dataCollectionService := service.NewDataCollectionServiceWithDB(db)

	// Create a test application first
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "datacollection",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		request       *service.CreateDataCollectionRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid data collection creation",
			request: &service.CreateDataCollectionRequest{
				AppID:       "app-001",
				Name:        "Test Collection",
				Description: "Test data collection",
				Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
				Config:      `{"allowDuplicates": false}`,
				CreatedBy:   "user-001",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			collection, err := dataCollectionService.CreateDataCollection(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, collection)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, collection)
				assert.NotEmpty(t, collection.ID)
				assert.Equal(t, tt.request.AppID, collection.AppID)
				assert.Equal(t, tt.request.Name, collection.Name)
				assert.Equal(t, tt.request.Description, collection.Description)
				assert.Equal(t, tt.request.Schema, collection.Schema)
				assert.Equal(t, tt.request.Config, collection.Config)
				assert.True(t, collection.IsActive)
				assert.Equal(t, tt.request.CreatedBy, collection.CreatedBy)
				assert.WithinDuration(t, time.Now(), collection.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), collection.UpdatedAt, time.Second)
			}
		})
	}

	// Test case for non-existent application
	t.Run("Non-existent application", func(t *testing.T) {
		request := &service.CreateDataCollectionRequest{
			AppID:       "non-existent-app",
			Name:        "Orphan Collection",
			Description: "Orphan data collection",
			Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
			Config:      `{"allowDuplicates": false}`,
			CreatedBy:   "user-001",
		}

		// Execute
		collection, err := dataCollectionService.CreateDataCollection(context.Background(), request)

		// Assert
		assert.NoError(t, err) // Note: This doesn't validate app existence in the service
		assert.NotNil(t, collection)
		assert.NotEmpty(t, collection.ID)
		assert.Equal(t, request.AppID, collection.AppID)
		assert.Equal(t, request.Name, collection.Name)
		assert.Equal(t, request.Description, collection.Description)
		assert.Equal(t, request.Schema, collection.Schema)
		assert.Equal(t, request.Config, collection.Config)
		assert.True(t, collection.IsActive)
		assert.Equal(t, request.CreatedBy, collection.CreatedBy)
		assert.WithinDuration(t, time.Now(), collection.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), collection.UpdatedAt, time.Second)
	})
}

func TestDataCollectionService_UpdateDataCollection(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.DataCollection{}, &domain.DataCollectionEntry{}, &domain.Application{})

	dataCollectionService := service.NewDataCollectionServiceWithDB(db)

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "datacollection",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a data collection for testing
	createReq := &service.CreateDataCollectionRequest{
		AppID:       "app-001",
		Name:        "Original Collection",
		Description: "Original description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"allowDuplicates": false}`,
		CreatedBy:   "user-001",
	}
	createdCollection, err := dataCollectionService.CreateDataCollection(context.Background(), createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdCollection)

	// Test cases
	tests := []struct {
		name          string
		collectionID  string
		request       *service.UpdateDataCollectionRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name:         "Valid data collection update",
			collectionID: createdCollection.ID,
			request: &service.UpdateDataCollectionRequest{
				Name:        "Updated Collection",
				Description: "Updated description",
				Schema:      `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "number"}}}`,
				Config:      `{"allowDuplicates": true}`,
				IsActive:    boolPtr(false),
			},
			expectError: false,
		},
		{
			name:          "Update non-existent data collection",
			collectionID:  "non-existent-id",
			request:       &service.UpdateDataCollectionRequest{Name: "Updated Name"},
			expectError:   true,
			errorMessage:  "data collection not found",
		},
		{
			name:         "Partial update - name only",
			collectionID: createdCollection.ID,
			request: &service.UpdateDataCollectionRequest{
				Name: "Name Only Update",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := dataCollectionService.UpdateDataCollection(context.Background(), tt.collectionID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify the update
				updatedCollection, getErr := dataCollectionService.GetDataCollection(context.Background(), tt.collectionID)
				assert.NoError(t, getErr)
				assert.NotNil(t, updatedCollection)

				// Check updated fields
				if tt.request.Name != "" {
					assert.Equal(t, tt.request.Name, updatedCollection.Name)
				}
				if tt.request.Description != "" {
					assert.Equal(t, tt.request.Description, updatedCollection.Description)
				}
				if tt.request.Schema != "" {
					assert.Equal(t, tt.request.Schema, updatedCollection.Schema)
				}
				if tt.request.Config != "" {
					assert.Equal(t, tt.request.Config, updatedCollection.Config)
				}
				if tt.request.IsActive != nil {
					assert.Equal(t, *tt.request.IsActive, updatedCollection.IsActive)
				}
				// UpdatedAt should be changed
				assert.True(t, updatedCollection.UpdatedAt.After(createdCollection.UpdatedAt))
			}
		})
	}
}

func TestDataCollectionService_DeleteDataCollection(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.DataCollection{}, &domain.DataCollectionEntry{}, &domain.Application{})

	dataCollectionService := service.NewDataCollectionServiceWithDB(db)

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "datacollection",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a data collection for testing
	createReq := &service.CreateDataCollectionRequest{
		AppID:       "app-001",
		Name:        "Test Collection",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"allowDuplicates": false}`,
		CreatedBy:   "user-001",
	}
	createdCollection, err := dataCollectionService.CreateDataCollection(context.Background(), createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdCollection)

	// Test cases
	tests := []struct {
		name          string
		collectionID  string
		expectError   bool
		errorMessage  string
	}{
		{
			name:         "Delete non-existent data collection",
			collectionID: "non-existent-id",
			expectError:  true,
			errorMessage: "data collection not found",
		},
		{
			name:         "Valid data collection deletion",
			collectionID: createdCollection.ID,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := dataCollectionService.DeleteDataCollection(context.Background(), tt.collectionID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify data collection is deleted
				_, getErr := dataCollectionService.GetDataCollection(context.Background(), tt.collectionID)
				assert.Error(t, getErr)
				assert.Equal(t, "data collection not found", getErr.Error())
			}
		})
	}
}

func TestDataCollectionService_ListDataCollections(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.DataCollection{}, &domain.DataCollectionEntry{}, &domain.Application{})

	dataCollectionService := service.NewDataCollectionServiceWithDB(db)

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "datacollection",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create test data collections
	collections := []*service.CreateDataCollectionRequest{
		{
			AppID:       "app-001",
			Name:        "Collection 1",
			Description: "First collection",
			Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
			Config:      `{"allowDuplicates": false}`,
			CreatedBy:   "user-001",
		},
		{
			AppID:       "app-001",
			Name:        "Collection 2",
			Description: "Second collection",
			Schema:      `{"type": "object", "properties": {"age": {"type": "number"}}}`,
			Config:      `{"allowDuplicates": true}`,
			CreatedBy:   "user-001",
		},
		{
			AppID:       "app-001",
			Name:        "Collection 3",
			Description: "Third collection",
			Schema:      `{"type": "object", "properties": {"email": {"type": "string"}}}`,
			Config:      `{"allowDuplicates": false}`,
			CreatedBy:   "user-001",
		},
	}

	for _, collectionReq := range collections {
		_, err := dataCollectionService.CreateDataCollection(context.Background(), collectionReq)
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
			result, total, err := dataCollectionService.ListDataCollections(context.Background(), tt.appID, tt.page, tt.size)

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

func TestDataCollectionService_GetDataCollection(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.DataCollection{}, &domain.DataCollectionEntry{}, &domain.Application{})

	dataCollectionService := service.NewDataCollectionServiceWithDB(db)

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "datacollection",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a data collection for testing
	createReq := &service.CreateDataCollectionRequest{
		AppID:       "app-001",
		Name:        "Test Collection",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"allowDuplicates": false}`,
		CreatedBy:   "user-001",
	}
	createdCollection, err := dataCollectionService.CreateDataCollection(context.Background(), createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdCollection)

	// Test cases
	tests := []struct {
		name          string
		collectionID  string
		expectError   bool
		errorMessage  string
	}{
		{
			name:         "Get non-existent data collection",
			collectionID: "non-existent-id",
			expectError:  true,
			errorMessage: "data collection not found",
		},
		{
			name:         "Get existing data collection",
			collectionID: createdCollection.ID,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := dataCollectionService.GetDataCollection(context.Background(), tt.collectionID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, createdCollection.ID, result.ID)
				assert.Equal(t, createdCollection.AppID, result.AppID)
				assert.Equal(t, createdCollection.Name, result.Name)
				assert.Equal(t, createdCollection.Description, result.Description)
				assert.Equal(t, createdCollection.Schema, result.Schema)
				assert.Equal(t, createdCollection.Config, result.Config)
				assert.Equal(t, createdCollection.IsActive, result.IsActive)
				assert.Equal(t, createdCollection.CreatedBy, result.CreatedBy)
				assert.Equal(t, createdCollection.CreatedAt.Unix(), result.CreatedAt.Unix())
				assert.Equal(t, createdCollection.UpdatedAt.Unix(), result.UpdatedAt.Unix())
			}
		})
	}
}

func TestDataCollectionService_SubmitDataEntry(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.DataCollection{}, &domain.DataCollectionEntry{}, &domain.Application{})

	dataCollectionService := service.NewDataCollectionServiceWithDB(db)

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "datacollection",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a data collection for testing
	createReq := &service.CreateDataCollectionRequest{
		AppID:       "app-001",
		Name:        "Test Collection",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"allowDuplicates": false}`,
		CreatedBy:   "user-001",
	}
	createdCollection, err := dataCollectionService.CreateDataCollection(context.Background(), createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdCollection)

	// Test cases
	tests := []struct {
		name          string
		request       *service.SubmitDataEntryRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid data entry submission",
			request: &service.SubmitDataEntryRequest{
				CollectionID: createdCollection.ID,
				Data:         `{"name": "John Doe"}`,
				CreatedBy:    "user-001",
			},
			expectError: false,
		},
		{
			name: "Submit to non-existent collection",
			request: &service.SubmitDataEntryRequest{
				CollectionID: "non-existent-id",
				Data:         `{"name": "Jane Doe"}`,
				CreatedBy:    "user-001",
			},
			expectError: true,
			errorMessage: "data collection not found or inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			entry, err := dataCollectionService.SubmitDataEntry(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, entry)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entry)
				assert.NotEmpty(t, entry.ID)
				assert.Equal(t, tt.request.CollectionID, entry.CollectionID)
				assert.Equal(t, tt.request.Data, entry.Data)
				assert.Equal(t, tt.request.CreatedBy, entry.CreatedBy)
				assert.WithinDuration(t, time.Now(), entry.CreatedAt, time.Second)
			}
		})
	}
}

func TestDataCollectionService_ListDataEntries(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.DataCollection{}, &domain.DataCollectionEntry{}, &domain.Application{})

	dataCollectionService := service.NewDataCollectionServiceWithDB(db)

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "datacollection",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create a data collection for testing
	createReq := &service.CreateDataCollectionRequest{
		AppID:       "app-001",
		Name:        "Test Collection",
		Description: "Test description",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		Config:      `{"allowDuplicates": false}`,
		CreatedBy:   "user-001",
	}
	createdCollection, err := dataCollectionService.CreateDataCollection(context.Background(), createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdCollection)

	// Create test data entries
	entries := []*service.SubmitDataEntryRequest{
		{
			CollectionID: createdCollection.ID,
			Data:         `{"name": "John Doe"}`,
			CreatedBy:    "user-001",
		},
		{
			CollectionID: createdCollection.ID,
			Data:         `{"name": "Jane Smith"}`,
			CreatedBy:    "user-002",
		},
		{
			CollectionID: createdCollection.ID,
			Data:         `{"name": "Bob Johnson"}`,
			CreatedBy:    "user-003",
		},
	}

	for _, entryReq := range entries {
		_, err := dataCollectionService.SubmitDataEntry(context.Background(), entryReq)
		assert.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name              string
		collectionID      string
		page              int
		size              int
		expectedCount     int
		totalCount        int64
		expectError       bool
		errorMessage      string
	}{
		{
			name:          "List first page",
			collectionID:  createdCollection.ID,
			page:          1,
			size:          2,
			expectedCount: 2,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List second page",
			collectionID:  createdCollection.ID,
			page:          2,
			size:          2,
			expectedCount: 1,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List with large page size",
			collectionID:  createdCollection.ID,
			page:          1,
			size:          10,
			expectedCount: 3,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List with zero page",
			collectionID:  createdCollection.ID,
			page:          0,
			size:          10,
			expectedCount: 3,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List non-existent collection",
			collectionID:  "non-existent-id",
			page:          1,
			size:          10,
			expectedCount: 0,
			totalCount:    0,
			expectError:   true,
			errorMessage:  "data collection not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, total, err := dataCollectionService.ListDataEntries(context.Background(), tt.collectionID, tt.page, tt.size)

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
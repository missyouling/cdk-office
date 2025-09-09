package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/document/domain"
	"cdk-office/internal/shared/testutils"
)

// TestCategoryService tests the CategoryService
func TestCategoryService(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create category service with database connection
	categoryService := &CategoryService{
		db: testDB,
	}

	// Test CreateCategory
	t.Run("CreateCategory", func(t *testing.T) {
		ctx := context.Background()

		category, err := categoryService.CreateCategory(ctx, "Test Category", "A test category", "")

		assert.NoError(t, err)
		assert.NotNil(t, category)
		assert.Equal(t, "Test Category", category.Name)
		assert.Equal(t, "A test category", category.Description)
		assert.Equal(t, "", category.ParentID)
	})

	// Test GetCategory
	t.Run("GetCategory", func(t *testing.T) {
		ctx := context.Background()

		// First create a category
		createdCategory, err := categoryService.CreateCategory(ctx, "Get Test Category", "A category to get", "")
		assert.NoError(t, err)
		assert.NotNil(t, createdCategory)

		// Now get the category
		retrievedCategory, err := categoryService.GetCategory(ctx, createdCategory.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedCategory)
		assert.Equal(t, createdCategory.ID, retrievedCategory.ID)
		assert.Equal(t, createdCategory.Name, retrievedCategory.Name)
	})

	// Test UpdateCategory
	t.Run("UpdateCategory", func(t *testing.T) {
		ctx := context.Background()

		// First create a category
		category, err := categoryService.CreateCategory(ctx, "Update Test Category", "A category to update", "")
		assert.NoError(t, err)
		assert.NotNil(t, category)

		// Now update the category
		err = categoryService.UpdateCategory(ctx, category.ID, "Updated Category", "An updated category")
		assert.NoError(t, err)

		// Verify the update
		updatedCategory, err := categoryService.GetCategory(ctx, category.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedCategory)
		assert.Equal(t, "Updated Category", updatedCategory.Name)
		assert.Equal(t, "An updated category", updatedCategory.Description)
	})

	// Test DeleteCategory
	t.Run("DeleteCategory", func(t *testing.T) {
		ctx := context.Background()

		// First create a category
		category, err := categoryService.CreateCategory(ctx, "Delete Test Category", "A category to delete", "")
		assert.NoError(t, err)
		assert.NotNil(t, category)

		// Now delete the category
		err = categoryService.DeleteCategory(ctx, category.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = categoryService.GetCategory(ctx, category.ID)
		assert.Error(t, err)
		assert.Equal(t, "category not found", err.Error())
	})

	// Test ListCategories
	t.Run("ListCategories", func(t *testing.T) {
		ctx := context.Background()

		// Create a few categories
		categoryNames := []string{"Category 1", "Category 2", "Category 3"}
		for _, name := range categoryNames {
			_, err := categoryService.CreateCategory(ctx, name, "A test category", "")
			assert.NoError(t, err)
		}

		// List categories
		categories, err := categoryService.ListCategories(ctx, "")
		assert.NoError(t, err)
		assert.NotNil(t, categories)
		assert.GreaterOrEqual(t, len(categories), 3)
	})

	// Test AssignDocumentToCategory and related functionality
	t.Run("AssignAndListDocumentCategories", func(t *testing.T) {
		ctx := context.Background()

		// Create a category
		category, err := categoryService.CreateCategory(ctx, "Document Category", "A category for documents", "")
		assert.NoError(t, err)
		assert.NotNil(t, category)

		// Create a document (using the test database directly)
	document := &domain.Document{
		ID:       "doc_" + time.Now().Format("20060102150405") + fmt.Sprintf("%d", time.Now().Nanosecond()),
		Title:    "Test Document",
		FilePath: "/path/to/document.txt",
		FileSize: 1024,
		MimeType: "text/plain",
		OwnerID:  "user_123",
		TeamID:   "team_123",
		Status:   "active",
	}

	err = testDB.Create(document).Error
	assert.NoError(t, err)

		// Assign document to category
		err = categoryService.AssignDocumentToCategory(ctx, document.ID, category.ID)
		assert.NoError(t, err)

		// Get document categories
		categories, err := categoryService.GetDocumentCategories(ctx, document.ID)
		assert.NoError(t, err)
		assert.NotNil(t, categories)
		assert.Len(t, categories, 1)
		assert.Equal(t, category.ID, categories[0].ID)
	})

	// Test RemoveDocumentFromCategory
	t.Run("RemoveDocumentFromCategory", func(t *testing.T) {
		ctx := context.Background()

		// Create a category
		category, err := categoryService.CreateCategory(ctx, "Remove Document Category", "A category for document removal", "")
		assert.NoError(t, err)
		assert.NotNil(t, category)

		// Create a document (using the test database directly)
	document := &domain.Document{
		ID:       "doc_" + time.Now().Format("20060102150405") + fmt.Sprintf("%d", time.Now().Nanosecond()),
		Title:    "Test Document for Removal",
		FilePath: "/path/to/document.txt",
		FileSize: 1024,
		MimeType: "text/plain",
		OwnerID:  "user_123",
		TeamID:   "team_123",
		Status:   "active",
	}

	err = testDB.Create(document).Error
	assert.NoError(t, err)

		// Assign document to category
		err = categoryService.AssignDocumentToCategory(ctx, document.ID, category.ID)
		assert.NoError(t, err)

		// Verify assignment
		categories, err := categoryService.GetDocumentCategories(ctx, document.ID)
		assert.NoError(t, err)
		assert.NotNil(t, categories)
		assert.Len(t, categories, 1)

		// Remove document from category
		err = categoryService.RemoveDocumentFromCategory(ctx, document.ID, category.ID)
		assert.NoError(t, err)

		// Verify removal
		categories, err = categoryService.GetDocumentCategories(ctx, document.ID)
		assert.NoError(t, err)
		assert.NotNil(t, categories)
		assert.Len(t, categories, 0)
	})
}

// TestCategoryServiceAdditional tests additional scenarios for the CategoryService
func TestCategoryServiceAdditional(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create category service with database connection
	categoryService := &CategoryService{
		db: testDB,
	}

	// Test CreateCategory with parent
	t.Run("CreateCategoryWithParent", func(t *testing.T) {
		ctx := context.Background()

		// First create a parent category
		parentCategory, err := categoryService.CreateCategory(ctx, "Parent Category", "A parent category", "")
		assert.NoError(t, err)
		assert.NotNil(t, parentCategory)

		// Create a child category
		childCategory, err := categoryService.CreateCategory(ctx, "Child Category", "A child category", parentCategory.ID)
		assert.NoError(t, err)
		assert.NotNil(t, childCategory)
		assert.Equal(t, parentCategory.ID, childCategory.ParentID)
	})

	// Test CreateCategory with non-existent parent
	t.Run("CreateCategoryWithNonExistentParent", func(t *testing.T) {
		ctx := context.Background()

		category, err := categoryService.CreateCategory(ctx, "Invalid Parent Category", "A category with invalid parent", "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, category)
		assert.Equal(t, "parent category not found", err.Error())
	})

	// Test GetCategory with non-existent ID
	t.Run("GetCategoryNotFound", func(t *testing.T) {
		ctx := context.Background()

		category, err := categoryService.GetCategory(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, category)
		assert.Equal(t, "category not found", err.Error())
	})

	// Test UpdateCategory with non-existent ID
	t.Run("UpdateCategoryNotFound", func(t *testing.T) {
		ctx := context.Background()

		err := categoryService.UpdateCategory(ctx, "non-existent-id", "Updated Category", "An updated category")

		assert.Error(t, err)
		assert.Equal(t, "category not found", err.Error())
	})

	// Test DeleteCategory with non-existent ID
	t.Run("DeleteCategoryNotFound", func(t *testing.T) {
		ctx := context.Background()

		err := categoryService.DeleteCategory(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Equal(t, "category not found", err.Error())
	})

	// Test DeleteCategory with child categories
	t.Run("DeleteCategoryWithChildren", func(t *testing.T) {
		ctx := context.Background()

		// First create a parent category
	parentCategory, err := categoryService.CreateCategory(ctx, "Parent Category with Children "+fmt.Sprintf("%d", time.Now().Nanosecond()), "A parent category with children", "")
	assert.NoError(t, err)
	assert.NotNil(t, parentCategory)

	// Create a child category
	_, err = categoryService.CreateCategory(ctx, "Child Category "+fmt.Sprintf("%d", time.Now().Nanosecond()), "A child category", parentCategory.ID)
	assert.NoError(t, err)

	// Try to delete the parent category
	err = categoryService.DeleteCategory(ctx, parentCategory.ID)

	assert.Error(t, err)
	assert.Equal(t, "cannot delete category with child categories", err.Error())
	})

	// Test ListCategories with parent ID
	t.Run("ListCategoriesWithParent", func(t *testing.T) {
		ctx := context.Background()

		// First create a parent category
		parentCategory, err := categoryService.CreateCategory(ctx, "Parent Category for List", "A parent category for listing", "")
		assert.NoError(t, err)
		assert.NotNil(t, parentCategory)

		// Create a few child categories
	childNames := []string{"Child 1", "Child 2", "Child 3"}
	for i, name := range childNames {
		_, err := categoryService.CreateCategory(ctx, name, "A child category "+fmt.Sprintf("%d", i), parentCategory.ID)
		assert.NoError(t, err)
	}

		// List child categories
		categories, err := categoryService.ListCategories(ctx, parentCategory.ID)
		assert.NoError(t, err)
		assert.NotNil(t, categories)
		assert.Len(t, categories, 3)
	})

	// Test AssignDocumentToCategory with non-existent document
	t.Run("AssignDocumentToCategoryDocumentNotFound", func(t *testing.T) {
		ctx := context.Background()

		// Create a category
		category, err := categoryService.CreateCategory(ctx, "Assign Category", "A category for assignment", "")
		assert.NoError(t, err)
		assert.NotNil(t, category)

		// Try to assign a non-existent document to the category
		err = categoryService.AssignDocumentToCategory(ctx, "non-existent-id", category.ID)

		assert.Error(t, err)
		assert.Equal(t, "document not found", err.Error())
	})

	// Test AssignDocumentToCategory with non-existent category
	t.Run("AssignDocumentToCategoryCategoryNotFound", func(t *testing.T) {
		ctx := context.Background()

		// Create a document first
		document := &domain.Document{
			ID:       "doc_" + fmt.Sprintf("%d", time.Now().Nanosecond()),
			Title:    "Test Document",
			FilePath: "/path/to/document.txt",
			FileSize: 1024,
			MimeType: "text/plain",
			OwnerID:  "user_123",
			TeamID:   "team_123",
			Status:   "active",
		}

		err := testDB.Create(document).Error
		assert.NoError(t, err)

		// Try to assign a document to a non-existent category
		err = categoryService.AssignDocumentToCategory(ctx, document.ID, "non-existent-id")

		assert.Error(t, err)
		assert.Equal(t, "category not found", err.Error())
	})

	// Test RemoveDocumentFromCategory with non-existent relation
	t.Run("RemoveDocumentFromCategoryNotFound", func(t *testing.T) {
		ctx := context.Background()

		// Try to remove a document from a category when no relation exists
		err := categoryService.RemoveDocumentFromCategory(ctx, "doc_123", "cat_123")

		assert.Error(t, err)
		assert.Equal(t, "document-category relation not found", err.Error())
	})

	// Test GetDocumentCategories with non-existent document
	t.Run("GetDocumentCategoriesDocumentNotFound", func(t *testing.T) {
		ctx := context.Background()

		// Try to get categories for a non-existent document
		categories, err := categoryService.GetDocumentCategories(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, categories)
		assert.Equal(t, "document not found", err.Error())
	})

	// Test multiple category operations
	t.Run("MultipleCategoryOperations", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple categories
		categoryNames := []string{"Multi Test 1", "Multi Test 2", "Multi Test 3"}
		var createdCategories []*domain.DocumentCategory
		for _, name := range categoryNames {
			category, err := categoryService.CreateCategory(ctx, name, "A multi test category", "")
			assert.NoError(t, err)
			assert.NotNil(t, category)
			createdCategories = append(createdCategories, category)
		}

		// Update all categories
		for _, category := range createdCategories {
			err := categoryService.UpdateCategory(ctx, category.ID, "Updated "+category.Name, "An updated category")
			assert.NoError(t, err)
		}

		// Verify updates
		for _, category := range createdCategories {
			updatedCategory, err := categoryService.GetCategory(ctx, category.ID)
			assert.NoError(t, err)
			assert.NotNil(t, updatedCategory)
			assert.Contains(t, updatedCategory.Name, "Updated")
		}

		// Delete all categories
		for _, category := range createdCategories {
			err := categoryService.DeleteCategory(ctx, category.ID)
			assert.NoError(t, err)
		}

		// Verify deletions
		for _, category := range createdCategories {
			_, err := categoryService.GetCategory(ctx, category.ID)
			assert.Error(t, err)
			assert.Equal(t, "category not found", err.Error())
		}
	})
}
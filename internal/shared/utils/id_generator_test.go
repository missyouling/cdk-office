package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAppID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateAppID()
	id2 := GenerateAppID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateAppID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "app_"))
	assert.True(t, strings.HasPrefix(id2, "app_"))
	assert.True(t, strings.HasPrefix(id3, "app_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateQRCodeID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateQRCodeID()
	id2 := GenerateQRCodeID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateQRCodeID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "qrcode_"))
	assert.True(t, strings.HasPrefix(id2, "qrcode_"))
	assert.True(t, strings.HasPrefix(id3, "qrcode_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateBatchID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateBatchID()
	id2 := GenerateBatchID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateBatchID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "batch_"))
	assert.True(t, strings.HasPrefix(id2, "batch_"))
	assert.True(t, strings.HasPrefix(id3, "batch_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateFormID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateFormID()
	id2 := GenerateFormID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateFormID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "form_"))
	assert.True(t, strings.HasPrefix(id2, "form_"))
	assert.True(t, strings.HasPrefix(id3, "form_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateFormDesignID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateFormDesignID()
	id2 := GenerateFormDesignID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateFormDesignID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "form_design_"))
	assert.True(t, strings.HasPrefix(id2, "form_design_"))
	assert.True(t, strings.HasPrefix(id3, "form_design_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGeneratePermissionID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GeneratePermissionID()
	id2 := GeneratePermissionID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GeneratePermissionID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "perm_"))
	assert.True(t, strings.HasPrefix(id2, "perm_"))
	assert.True(t, strings.HasPrefix(id3, "perm_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateUserPermissionID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateUserPermissionID()
	id2 := GenerateUserPermissionID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateUserPermissionID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "user_perm_"))
	assert.True(t, strings.HasPrefix(id2, "user_perm_"))
	assert.True(t, strings.HasPrefix(id3, "user_perm_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateDataCollectionID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateDataCollectionID()
	id2 := GenerateDataCollectionID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateDataCollectionID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "data_collection_"))
	assert.True(t, strings.HasPrefix(id2, "data_collection_"))
	assert.True(t, strings.HasPrefix(id3, "data_collection_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateDataCollectionEntryID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateDataCollectionEntryID()
	id2 := GenerateDataCollectionEntryID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateDataCollectionEntryID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "data_entry_"))
	assert.True(t, strings.HasPrefix(id2, "data_entry_"))
	assert.True(t, strings.HasPrefix(id3, "data_entry_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateEmployeeID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateEmployeeID()
	id2 := GenerateEmployeeID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateEmployeeID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "emp_"))
	assert.True(t, strings.HasPrefix(id2, "emp_"))
	assert.True(t, strings.HasPrefix(id3, "emp_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateContractID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateContractID()
	id2 := GenerateContractID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateContractID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "contract_"))
	assert.True(t, strings.HasPrefix(id2, "contract_"))
	assert.True(t, strings.HasPrefix(id3, "contract_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateModuleID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateModuleID()
	id2 := GenerateModuleID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateModuleID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "module_"))
	assert.True(t, strings.HasPrefix(id2, "module_"))
	assert.True(t, strings.HasPrefix(id3, "module_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGeneratePluginID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GeneratePluginID()
	id2 := GeneratePluginID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GeneratePluginID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "plugin_"))
	assert.True(t, strings.HasPrefix(id2, "plugin_"))
	assert.True(t, strings.HasPrefix(id3, "plugin_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateSurveyID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateSurveyID()
	id2 := GenerateSurveyID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateSurveyID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "survey_"))
	assert.True(t, strings.HasPrefix(id2, "survey_"))
	assert.True(t, strings.HasPrefix(id3, "survey_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateBusinessPermissionID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateBusinessPermissionID()
	id2 := GenerateBusinessPermissionID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateBusinessPermissionID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "biz_perm_"))
	assert.True(t, strings.HasPrefix(id2, "biz_perm_"))
	assert.True(t, strings.HasPrefix(id3, "biz_perm_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateDepartmentID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateDepartmentID()
	id2 := GenerateDepartmentID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateDepartmentID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "dept_"))
	assert.True(t, strings.HasPrefix(id2, "dept_"))
	assert.True(t, strings.HasPrefix(id3, "dept_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateLifecycleID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateLifecycleID()
	id2 := GenerateLifecycleID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateLifecycleID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "lifecycle_"))
	assert.True(t, strings.HasPrefix(id2, "lifecycle_"))
	assert.True(t, strings.HasPrefix(id3, "lifecycle_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateDocumentID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateDocumentID()
	id2 := GenerateDocumentID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateDocumentID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "doc_"))
	assert.True(t, strings.HasPrefix(id2, "doc_"))
	assert.True(t, strings.HasPrefix(id3, "doc_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateDocumentVersionID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateDocumentVersionID()
	id2 := GenerateDocumentVersionID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateDocumentVersionID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "doc_ver_"))
	assert.True(t, strings.HasPrefix(id2, "doc_ver_"))
	assert.True(t, strings.HasPrefix(id3, "doc_ver_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateSurveyResponseID(t *testing.T) {
	// Generate two IDs to ensure they are unique
	id1 := GenerateSurveyResponseID()
	id2 := GenerateSurveyResponseID()
	
	// Add a small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	
	id3 := GenerateSurveyResponseID()
	
	// Assertions
	assert.True(t, strings.HasPrefix(id1, "survey_resp_"))
	assert.True(t, strings.HasPrefix(id2, "survey_resp_"))
	assert.True(t, strings.HasPrefix(id3, "survey_resp_"))
	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id2, id3)
	assert.NotEqual(t, id1, id3)
	
	// Check that the ID contains a timestamp
	assert.Contains(t, id1, time.Now().Format("20060102"))
}

func TestGenerateRandomSuffix(t *testing.T) {
	// Generate multiple suffixes to ensure they are different
	suffix1 := generateRandomSuffix()
	suffix2 := generateRandomSuffix()
	
	// Add a small delay
	time.Sleep(time.Millisecond)
	
	suffix3 := generateRandomSuffix()
	
	// Assertions
	assert.Len(t, suffix1, 6)
	assert.Len(t, suffix2, 6)
	assert.Len(t, suffix3, 6)
	// Note: There's a small chance these could be equal, but it's very unlikely
	// We're not asserting they're different because randomness can occasionally produce the same value
}
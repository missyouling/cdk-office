package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateAppID generates a unique ID for applications
func GenerateAppID() string {
	// In a real application, use a proper ID generation library like uuid
	return "app_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateQRCodeID generates a unique ID for QR codes
func GenerateQRCodeID() string {
	// In a real application, use a proper ID generation library like uuid
	return "qrcode_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateBatchID generates a unique ID for batch QR codes
func GenerateBatchID() string {
	// In a real application, use a proper ID generation library like uuid
	return "batch_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateFormID generates a unique ID for forms
func GenerateFormID() string {
	// In a real application, use a proper ID generation library like uuid
	return "form_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateFormDesignID generates a unique ID for form designs
func GenerateFormDesignID() string {
	// In a real application, use a proper ID generation library like uuid
	return "form_design_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GeneratePermissionID generates a unique ID for permissions
func GeneratePermissionID() string {
	// In a real application, use a proper ID generation library like uuid
	return "perm_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateUserPermissionID generates a unique ID for user permissions
func GenerateUserPermissionID() string {
	// In a real application, use a proper ID generation library like uuid
	return "user_perm_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateDataCollectionID generates a unique ID for data collections
func GenerateDataCollectionID() string {
	// In a real application, use a proper ID generation library like uuid
	return "data_collection_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateDataCollectionEntryID generates a unique ID for data collection entries
func GenerateDataCollectionEntryID() string {
	// In a real application, use a proper ID generation library like uuid
	return "data_entry_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateEmployeeID generates a unique ID for employees
func GenerateEmployeeID() string {
	// In a real application, use a proper ID generation library like uuid
	return "emp_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateContractID generates a unique ID for contracts
func GenerateContractID() string {
	// In a real application, use a proper ID generation library like uuid
	return "contract_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateModuleID generates a unique ID for modules
func GenerateModuleID() string {
	// In a real application, use a proper ID generation library like uuid
	return "module_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GeneratePluginID generates a unique ID for plugins
func GeneratePluginID() string {
	// In a real application, use a proper ID generation library like uuid
	return "plugin_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateSurveyID generates a unique ID for surveys
func GenerateSurveyID() string {
	// In a real application, use a proper ID generation library like uuid
	return "survey_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateBusinessPermissionID generates a unique ID for business permissions
func GenerateBusinessPermissionID() string {
	// In a real application, use a proper ID generation library like uuid
	return "biz_perm_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateDepartmentID generates a unique ID for departments
func GenerateDepartmentID() string {
	// In a real application, use a proper ID generation library like uuid
	return "dept_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateLifecycleID generates a unique ID for employee lifecycle events
func GenerateLifecycleID() string {
	// In a real application, use a proper ID generation library like uuid
	return "lifecycle_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateDocumentID generates a unique ID for documents
func GenerateDocumentID() string {
	// In a real application, use a proper ID generation library like uuid
	return "doc_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateDocumentVersionID generates a unique ID for document versions
func GenerateDocumentVersionID() string {
	// In a real application, use a proper ID generation library like uuid
	return "doc_ver_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// GenerateSurveyResponseID generates a unique ID for survey responses
func GenerateSurveyResponseID() string {
	// In a real application, use a proper ID generation library like uuid
	return "survey_resp_" + time.Now().Format("20060102150405") + generateRandomSuffix()
}

// generateRandomSuffix generates a random suffix to ensure uniqueness
func generateRandomSuffix() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
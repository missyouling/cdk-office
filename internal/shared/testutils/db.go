package testutils

import (
	appdomain "cdk-office/internal/app/domain"
	documentdomain "cdk-office/internal/document/domain"
	employeedomain "cdk-office/internal/employee/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB initializes a test database with all required tables
func SetupTestDB() *gorm.DB {
	// Create a SQLite database in memory for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&appdomain.Application{})
	db.AutoMigrate(&appdomain.QRCode{})
	db.AutoMigrate(&appdomain.AppPermission{})
	db.AutoMigrate(&appdomain.AppUserPermission{})
	db.AutoMigrate(&appdomain.BatchQRCode{})
	db.AutoMigrate(&appdomain.BatchQRCodeItem{})
	db.AutoMigrate(&appdomain.DataCollection{})
	db.AutoMigrate(&appdomain.DataCollectionEntry{})
	db.AutoMigrate(&appdomain.FormData{})
	db.AutoMigrate(&appdomain.FormDataEntry{})
	db.AutoMigrate(&appdomain.FormDesign{})
	db.AutoMigrate(&documentdomain.Document{})
	db.AutoMigrate(&documentdomain.DocumentVersion{})
	db.AutoMigrate(&documentdomain.DocumentCategory{})
	db.AutoMigrate(&documentdomain.DocumentCategoryRelation{})
	db.AutoMigrate(&employeedomain.Employee{})
	db.AutoMigrate(&employeedomain.Department{})
	db.AutoMigrate(&employeedomain.PerformanceReview{})
	db.AutoMigrate(&employeedomain.TerminationRecord{})
	db.AutoMigrate(&employeedomain.EmployeeSurvey{})
	db.AutoMigrate(&employeedomain.SurveyResponse{})
	db.AutoMigrate(&employeedomain.SurveyQuestion{})
	// Note: EmployeeLifecycleEvent is defined in service package, so we can't auto-migrate it here
	// We'll create the table manually
	db.Exec(`CREATE TABLE IF NOT EXISTS employee_lifecycle_events (
		id TEXT PRIMARY KEY,
		employee_id TEXT,
		event_type TEXT,
		old_value TEXT,
		new_value TEXT,
		effective_date DATETIME,
		reason TEXT,
		created_at DATETIME
	)`)

	// Initialize the database connection for testing
	database.InitDB(db)

	// Initialize logger for testing
	logger.InitTestLogger()

	return db
}
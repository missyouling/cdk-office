package database

import (
	"gorm.io/gorm"
)

var db *gorm.DB

// InitDB initializes the database connection
func InitDB(database *gorm.DB) {
	db = database
}

// GetDB returns the database connection
func GetDB() *gorm.DB {
	return db
}
package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetDatabaseConfig returns the database configuration from environment variables
func GetDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "cdk_office"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}
}

// InitDatabase initializes the database connection with connection pooling
func InitDatabase() *gorm.DB {
	// Check if DATABASE_URL environment variable is set
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		// Use DATABASE_URL for connection
		db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		// Get the underlying SQL database connection
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatal("Failed to get database instance:", err)
		}

		// Configure connection pool
		sqlDB.SetMaxIdleConns(10)           // Maximum number of idle connections
		sqlDB.SetMaxOpenConns(100)          // Maximum number of open connections
		sqlDB.SetConnMaxLifetime(0)         // Maximum amount of time a connection may be reused (0 means no limit)

		return db
	}

	// Fallback to individual environment variables
	config := GetDatabaseConfig()

	// Create connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	// Open database connection with connection pooling
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get the underlying SQL database connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)           // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)          // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(0)         // Maximum amount of time a connection may be reused (0 means no limit)

	return db
}

// getEnv returns the value of the environment variable or a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
)

func init() {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	// Create log file with current date
	logFile := filepath.Join(logDir, fmt.Sprintf("app_%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	// Create loggers
	infoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(file, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Info logs an info message
func Info(msg string, keysAndValues ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	logMsg := fmt.Sprintf("%s:%d %s", filepath.Base(file), line, msg)
	if len(keysAndValues) > 0 {
		logMsg += " "
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				logMsg += fmt.Sprintf("%v=%v ", keysAndValues[i], keysAndValues[i+1])
			} else {
				logMsg += fmt.Sprintf("%v=<?> ", keysAndValues[i])
			}
		}
	}
	infoLogger.Println(logMsg)
}

// Error logs an error message
func Error(msg string, keysAndValues ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	logMsg := fmt.Sprintf("%s:%d %s", filepath.Base(file), line, msg)
	if len(keysAndValues) > 0 {
		logMsg += " "
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				logMsg += fmt.Sprintf("%v=%v ", keysAndValues[i], keysAndValues[i+1])
			} else {
				logMsg += fmt.Sprintf("%v=<?> ", keysAndValues[i])
			}
		}
	}
	errorLogger.Println(logMsg)
}

// Warn logs a warning message
func Warn(msg string, keysAndValues ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	logMsg := fmt.Sprintf("%s:%d %s", filepath.Base(file), line, msg)
	if len(keysAndValues) > 0 {
		logMsg += " "
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				logMsg += fmt.Sprintf("%v=%v ", keysAndValues[i], keysAndValues[i+1])
			} else {
				logMsg += fmt.Sprintf("%v=<?> ", keysAndValues[i])
			}
		}
	}
	warnLogger.Println(logMsg)
}
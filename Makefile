# Makefile for running tests

# Test all modules
test:
	go test -v ./...

# Test app module
test-app:
	go test -v ./internal/app/service/test

# Test document module
test-document:
	go test -v ./internal/document/service/test

# Test employee module
test-employee:
	go test -v ./internal/employee/service/test

# Test dify module
test-dify:
	go test -v ./internal/dify/workflow/test

# Test with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Test with coverage for specific modules
test-app-coverage:
	go test -coverprofile=app_coverage.out ./internal/app/service/test
	go tool cover -html=app_coverage.out -o app_coverage.html

test-document-coverage:
	go test -coverprofile=document_coverage.out ./internal/document/service/test
	go tool cover -html=document_coverage.out -o document_coverage.html

test-employee-coverage:
	go test -coverprofile=employee_coverage.out ./internal/employee/service/test
	go tool cover -html=employee_coverage.out -o employee_coverage.html

test-dify-coverage:
	go test -coverprofile=dify_coverage.out ./internal/dify/workflow/test
	go tool cover -html=dify_coverage.out -o dify_coverage.html

.PHONY: test test-app test-document test-employee test-dify test-coverage test-app-coverage test-document-coverage test-employee-coverage test-dify-coverage
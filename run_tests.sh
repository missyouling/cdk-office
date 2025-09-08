#!/bin/bash

# Test script to run all tests and generate coverage reports

echo "Running tests for all modules..."

# Create coverage directory if it doesn't exist
mkdir -p coverage

# Run tests for each module and generate coverage reports
echo "Testing app module..."
go test -coverprofile=coverage/app.out ./internal/app/service/test
if [ $? -eq 0 ]; then
    echo "App module tests passed"
    go tool cover -html=coverage/app.out -o coverage/app.html
else
    echo "App module tests failed"
fi

echo "Testing document module..."
go test -coverprofile=coverage/document.out ./internal/document/service/test
if [ $? -eq 0 ]; then
    echo "Document module tests passed"
    go tool cover -html=coverage/document.out -o coverage/document.html
else
    echo "Document module tests failed"
fi

echo "Testing employee module..."
go test -coverprofile=coverage/employee.out ./internal/employee/service/test
if [ $? -eq 0 ]; then
    echo "Employee module tests passed"
    go tool cover -html=coverage/employee.out -o coverage/employee.html
else
    echo "Employee module tests failed"
fi

echo "Testing dify module..."
go test -coverprofile=coverage/dify.out ./internal/dify/workflow/test
if [ $? -eq 0 ]; then
    echo "Dify module tests passed"
    go tool cover -html=coverage/dify.out -o coverage/dify.html
else
    echo "Dify module tests failed"
fi

echo "Generating combined coverage report..."
echo "mode: set" > coverage/combined.out
tail -n +2 coverage/app.out >> coverage/combined.out 2>/dev/null
tail -n +2 coverage/document.out >> coverage/combined.out 2>/dev/null
tail -n +2 coverage/employee.out >> coverage/combined.out 2>/dev/null
tail -n +2 coverage/dify.out >> coverage/combined.out 2>/dev/null

go tool cover -html=coverage/combined.out -o coverage/combined.html

echo "Coverage reports generated in coverage directory"
echo "Open coverage/combined.html to view overall coverage"
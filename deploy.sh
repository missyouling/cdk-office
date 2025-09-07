#!/bin/bash

# CDK-Office Deployment Script

# Set script to exit on any error
set -e

# Load environment variables
if [ -f "deploy.env" ]; then
    source deploy.env
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    log_info "All dependencies are installed."
}

# Build the application
build_app() {
    log_info "Building the application..."
    
    # Build backend
    log_info "Building backend..."
    docker build -t cdk-office-backend .
    
    # Build frontend
    log_info "Building frontend..."
    cd frontend
    docker build -t cdk-office-frontend .
    cd ..
    
    log_info "Application built successfully."
}

# Start the application
start_app() {
    log_info "Starting the application..."
    
    # Start services with Docker Compose
    docker-compose up -d
    
    # Wait for services to start
    log_info "Waiting for services to start..."
    sleep 30
    
    # Check if services are running
    if docker-compose ps | grep -q "Up"; then
        log_info "Application started successfully."
    else
        log_error "Failed to start application. Check the logs for details."
        docker-compose logs
        exit 1
    fi
}

# Stop the application
stop_app() {
    log_info "Stopping the application..."
    
    # Stop services with Docker Compose
    docker-compose down
    
    log_info "Application stopped."
}

# Restart the application
restart_app() {
    log_info "Restarting the application..."
    
    stop_app
    start_app
    
    log_info "Application restarted."
}

# Show application status
status_app() {
    log_info "Application status:"
    
    # Show Docker Compose status
    docker-compose ps
}

# Show application logs
logs_app() {
    log_info "Application logs:"
    
    # Show Docker Compose logs
    docker-compose logs -f
}

# Create database backup
backup_db() {
    log_info "Creating database backup..."
    
    # Create backup directory if it doesn't exist
    mkdir -p $BACKUP_DIR
    
    # Create backup filename with timestamp
    BACKUP_FILE="$BACKUP_DIR/cdkoffice_backup_$(date +%Y%m%d_%H%M%S).sql"
    
    # Create database backup
    docker-compose exec postgres pg_dump -U $DB_USER $DB_NAME > $BACKUP_FILE
    
    log_info "Database backup created: $BACKUP_FILE"
}

# Restore database from backup
restore_db() {
    if [ -z "$1" ]; then
        log_error "Please provide a backup file to restore."
        exit 1
    fi
    
    BACKUP_FILE=$1
    
    if [ ! -f "$BACKUP_FILE" ]; then
        log_error "Backup file not found: $BACKUP_FILE"
        exit 1
    fi
    
    log_info "Restoring database from: $BACKUP_FILE"
    
    # Stop application
    stop_app
    
    # Restore database
    docker-compose exec -T postgres psql -U $DB_USER $DB_NAME < $BACKUP_FILE
    
    # Start application
    start_app
    
    log_info "Database restored successfully."
}

# Clean up old backups
cleanup_backups() {
    log_info "Cleaning up old backups..."
    
    # Remove backups older than BACKUP_RETENTION_DAYS
    find $BACKUP_DIR -name "cdkoffice_backup_*.sql" -mtime +$BACKUP_RETENTION_DAYS -delete
    
    log_info "Old backups cleaned up."
}

# Show help
show_help() {
    echo "CDK-Office Deployment Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build     Build the application"
    echo "  start     Start the application"
    echo "  stop      Stop the application"
    echo "  restart   Restart the application"
    echo "  status    Show application status"
    echo "  logs      Show application logs"
    echo "  backup    Create database backup"
    echo "  restore   Restore database from backup"
    echo "  cleanup   Clean up old backups"
    echo "  help      Show this help message"
    echo ""
}

# Main function
main() {
    # Check dependencies
    check_dependencies
    
    # Parse command line arguments
    case "$1" in
        build)
            build_app
            ;;
        start)
            start_app
            ;;
        stop)
            stop_app
            ;;
        restart)
            restart_app
            ;;
        status)
            status_app
            ;;
        logs)
            logs_app
            ;;
        backup)
            backup_db
            ;;
        restore)
            restore_db "$2"
            ;;
        cleanup)
            cleanup_backups
            ;;
        help)
            show_help
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
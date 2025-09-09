-- Create tables for CDK-Office application

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(20),
    password VARCHAR(255) NOT NULL,
    real_name VARCHAR(50),
    id_card VARCHAR(18),
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Employees table
CREATE TABLE IF NOT EXISTS employees (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) REFERENCES users(id),
    team_id VARCHAR(36),
    dept_id VARCHAR(36),
    employee_id VARCHAR(50) UNIQUE NOT NULL,
    real_name VARCHAR(50) NOT NULL,
    gender VARCHAR(10),
    birth_date DATE,
    hire_date DATE,
    position VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Departments table
CREATE TABLE IF NOT EXISTS departments (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    team_id VARCHAR(36),
    parent_id VARCHAR(36),
    level INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Role-Permission relationship table
CREATE TABLE IF NOT EXISTS role_permissions (
    id VARCHAR(36) PRIMARY KEY,
    role_id VARCHAR(36) REFERENCES roles(id),
    permission_id VARCHAR(36) REFERENCES permissions(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User-Role relationship table
CREATE TABLE IF NOT EXISTS user_roles (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) REFERENCES users(id),
    role VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Documents table
CREATE TABLE IF NOT EXISTS documents (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    file_path VARCHAR(500),
    file_size BIGINT,
    mime_type VARCHAR(100),
    team_id VARCHAR(36),
    created_by VARCHAR(36) REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Document versions table
CREATE TABLE IF NOT EXISTS document_versions (
    id VARCHAR(36) PRIMARY KEY,
    document_id VARCHAR(36) REFERENCES documents(id),
    file_path VARCHAR(500),
    file_size BIGINT,
    version_number INTEGER NOT NULL,
    created_by VARCHAR(36) REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Document categories table
CREATE TABLE IF NOT EXISTS document_categories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id VARCHAR(36),
    team_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Document-Category relationship table
CREATE TABLE IF NOT EXISTS document_category_relations (
    id VARCHAR(36) PRIMARY KEY,
    document_id VARCHAR(36) REFERENCES documents(id),
    category_id VARCHAR(36) REFERENCES document_categories(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- QR Codes table
CREATE TABLE IF NOT EXISTS qrcodes (
    id VARCHAR(36) PRIMARY KEY,
    content TEXT NOT NULL,
    file_path VARCHAR(500),
    team_id VARCHAR(36),
    created_by VARCHAR(36) REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Batch QR Codes table
CREATE TABLE IF NOT EXISTS batch_qrcodes (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    team_id VARCHAR(36),
    created_by VARCHAR(36) REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Batch QR Code items table
CREATE TABLE IF NOT EXISTS batch_qrcode_items (
    id VARCHAR(36) PRIMARY KEY,
    batch_id VARCHAR(36) REFERENCES batch_qrcodes(id),
    content TEXT NOT NULL,
    file_path VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Forms table
CREATE TABLE IF NOT EXISTS forms (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    schema JSONB,
    team_id VARCHAR(36),
    created_by VARCHAR(36) REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Form designs table
CREATE TABLE IF NOT EXISTS form_designs (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    schema JSONB,
    team_id VARCHAR(36),
    created_by VARCHAR(36) REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Data collections table
CREATE TABLE IF NOT EXISTS data_collections (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    schema JSONB,
    team_id VARCHAR(36),
    created_by VARCHAR(36) REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Data collection entries table
CREATE TABLE IF NOT EXISTS data_collection_entries (
    id VARCHAR(36) PRIMARY KEY,
    collection_id VARCHAR(36) REFERENCES data_collections(id),
    data JSONB,
    created_by VARCHAR(36) REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Employee lifecycle events table
CREATE TABLE IF NOT EXISTS employee_lifecycle_events (
    id VARCHAR(36) PRIMARY KEY,
    employee_id VARCHAR(36) REFERENCES employees(id),
    event_type VARCHAR(50) NOT NULL,
    old_value VARCHAR(255),
    new_value VARCHAR(255),
    effective_date DATE,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default roles
INSERT INTO roles (id, name, description) VALUES 
('role_admin', 'admin', 'System administrator with full access'),
('role_user', 'user', 'Regular user with standard access'),
('role_manager', 'manager', 'Manager with departmental access')
ON CONFLICT (name) DO NOTHING;

-- Insert default permissions
INSERT INTO permissions (id, name, resource, action, description) VALUES 
('perm_read_document', 'read_document', 'document', 'read', 'Read documents'),
('perm_write_document', 'write_document', 'document', 'write', 'Create and edit documents'),
('perm_delete_document', 'delete_document', 'document', 'delete', 'Delete documents'),
('perm_manage_user', 'manage_user', 'user', 'manage', 'Manage users'),
('perm_manage_role', 'manage_role', 'role', 'manage', 'Manage roles and permissions')
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to admin role
INSERT INTO role_permissions (id, role_id, permission_id) VALUES 
('rp_admin_1', 'role_admin', 'perm_read_document'),
('rp_admin_2', 'role_admin', 'perm_write_document'),
('rp_admin_3', 'role_admin', 'perm_delete_document'),
('rp_admin_4', 'role_admin', 'perm_manage_user'),
('rp_admin_5', 'role_admin', 'perm_manage_role')
ON CONFLICT DO NOTHING;
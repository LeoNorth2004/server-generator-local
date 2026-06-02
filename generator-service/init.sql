-- Generator Platform Database Initialization Script
-- PostgreSQL Schema for Operation Logs Persistence

CREATE DATABASE IF NOT EXISTS generator_platform;

\c generator_platform

-- Enable UUID extension (optional, for future use)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Operation Logs Table
-- Note: This table is auto-created by GORM AutoMigrate
-- This script is for reference and manual setup if needed

CREATE TABLE IF NOT EXISTS operation_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    username VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    resource_id INTEGER DEFAULT 0,
    details TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    error TEXT,
    ip_address VARCHAR(45),
    user_agent VARCHAR(255),
    duration_ms BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_action ON operation_logs(action);
CREATE INDEX IF NOT EXISTS idx_operation_logs_resource ON operation_logs(resource);
CREATE INDEX IF NOT EXISTS idx_operation_logs_status ON operation_logs(status);
CREATE INDEX IF NOT EXISTS idx_operation_logs_created_at ON operation_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_operation_logs_deleted_at ON operation_logs(deleted_at);

-- Projects Table
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    db_config TEXT,
    table_config TEXT,
    generated_code TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Comments
COMMENT ON TABLE operation_logs IS '操作日志表 - 记录所有用户操作';
COMMENT ON COLUMN operation_logs.action IS '操作类型: generate, regenerate, download, preview, login, create_project, update_project, delete_project, create_user, update_user, delete_user';
COMMENT ON COLUMN operation_logs.resource IS '资源类型: project, code, auth, user';
COMMENT ON COLUMN operation_logs.status IS '状态: success, failed, error';

COMMENT ON TABLE projects IS '项目表 - 存储生成的项目信息';

-- Grant permissions (adjust as needed)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO generator_service;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO generator_service;

SELECT 'Database initialization completed successfully!' AS result;

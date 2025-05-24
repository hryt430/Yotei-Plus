-- Task Management Database Initialization Script
CREATE DATABASE IF NOT EXISTS task_management CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE task_management;

-- Users table for authentication
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role ENUM('user', 'admin') DEFAULT 'user',
    email_verified BOOLEAN DEFAULT FALSE,
    last_login TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_username (username)
);

-- Refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(36) PRIMARY KEY,
    token VARCHAR(255) UNIQUE NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_token (token),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
);

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status ENUM('TODO', 'IN_PROGRESS', 'DONE') DEFAULT 'TODO',
    priority ENUM('LOW', 'MEDIUM', 'HIGH') DEFAULT 'MEDIUM',
    assignee_id VARCHAR(36) NULL,
    created_by VARCHAR(36) NOT NULL,
    due_date TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_assignee_id (assignee_id),
    INDEX idx_created_by (created_by),
    INDEX idx_due_date (due_date),
    INDEX idx_created_at (created_at),
    FULLTEXT idx_search (title, description)
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type ENUM('APP_NOTIFICATION', 'TASK_ASSIGNED', 'TASK_COMPLETED', 'TASK_DUE_SOON', 'SYSTEM_NOTICE') DEFAULT 'APP_NOTIFICATION',
    status ENUM('PENDING', 'SENT', 'READ', 'FAILED') DEFAULT 'PENDING',
    metadata JSON NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    sent_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_type (type),
    INDEX idx_created_at (created_at)
);

-- Task comments table (optional feature)
CREATE TABLE IF NOT EXISTS task_comments (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    comment TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_task_id (task_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);

-- Task attachments table (optional feature)
CREATE TABLE IF NOT EXISTS task_attachments (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    uploaded_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_task_id (task_id)
);

-- User roles table (for more complex role management)
CREATE TABLE IF NOT EXISTS user_roles (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    role_name VARCHAR(50) NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    granted_by VARCHAR(36) NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE KEY unique_user_role (user_id, role_name),
    INDEX idx_user_id (user_id),
    INDEX idx_role_name (role_name)
);

-- Insert sample data
INSERT INTO users (id, email, username, password, role, email_verified) VALUES 
('550e8400-e29b-41d4-a716-446655440000', 'admin@example.com', 'admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', TRUE),
('550e8400-e29b-41d4-a716-446655440001', 'user@example.com', 'testuser', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user', TRUE)
ON DUPLICATE KEY UPDATE email = VALUES(email);

-- Insert sample tasks
INSERT INTO tasks (id, title, description, status, priority, created_by) VALUES 
('660e8400-e29b-41d4-a716-446655440000', 'プロジェクト設計', 'アプリケーションアーキテクチャの設計', 'IN_PROGRESS', 'HIGH', '550e8400-e29b-41d4-a716-446655440000'),
('660e8400-e29b-41d4-a716-446655440001', 'データベース設計', 'ERD作成とテーブル設計', 'TODO', 'MEDIUM', '550e8400-e29b-41d4-a716-446655440000'),
('660e8400-e29b-41d4-a716-446655440002', 'API実装', 'REST API エンドポイントの実装', 'TODO', 'HIGH', '550e8400-e29b-41d4-a716-446655440001')
ON DUPLICATE KEY UPDATE title = VALUES(title);

-- Insert sample notifications
INSERT INTO notifications (id, user_id, title, message, type, status) VALUES 
('770e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', 'タスクが割り当てられました', '新しいタスク「API実装」が割り当てられました', 'TASK_ASSIGNED', 'SENT'),
('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'システムメンテナンス', 'システムメンテナンスが予定されています', 'SYSTEM_NOTICE', 'PENDING')
ON DUPLICATE KEY UPDATE title = VALUES(title);

-- Procedures for cleanup
DELIMITER //

CREATE PROCEDURE CleanupExpiredTokens()
BEGIN
    DELETE FROM refresh_tokens 
    WHERE expires_at < NOW() OR revoked_at IS NOT NULL;
END //

CREATE PROCEDURE GetTaskStatistics(IN user_id VARCHAR(36))
BEGIN
    SELECT 
        COUNT(*) as total_tasks,
        SUM(CASE WHEN status = 'TODO' THEN 1 ELSE 0 END) as todo_count,
        SUM(CASE WHEN status = 'IN_PROGRESS' THEN 1 ELSE 0 END) as in_progress_count,
        SUM(CASE WHEN status = 'DONE' THEN 1 ELSE 0 END) as done_count,
        SUM(CASE WHEN due_date < NOW() AND status != 'DONE' THEN 1 ELSE 0 END) as overdue_count
    FROM tasks 
    WHERE (assignee_id = user_id OR created_by = user_id);
END //

DELIMITER ;

-- Create indexes for better performance
CREATE INDEX idx_tasks_compound ON tasks (status, assignee_id, due_date);
CREATE INDEX idx_notifications_compound ON notifications (user_id, status, created_at);
CREATE INDEX idx_refresh_tokens_compound ON refresh_tokens (user_id, expires_at, revoked_at);

-- Update table statistics
ANALYZE TABLE users, tasks, notifications, refresh_tokens;

SELECT 'Database initialization completed successfully' as status;
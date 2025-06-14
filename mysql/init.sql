-- Task Management Database Initialization Script
CREATE DATABASE IF NOT EXISTS `Yotei-Plus` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Users table for authentication
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`users` (
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
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`refresh_tokens` (
    id VARCHAR(36) PRIMARY KEY,
    token VARCHAR(255) UNIQUE NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    INDEX idx_token (token),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
);

-- Tasks table
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`tasks` (
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
    FOREIGN KEY (assignee_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_assignee_id (assignee_id),
    INDEX idx_created_by (created_by),
    INDEX idx_due_date (due_date),
    INDEX idx_created_at (created_at),
    FULLTEXT idx_search (title, description)
);

-- Notifications table
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`notifications` (
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
    FOREIGN KEY (user_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_type (type),
    INDEX idx_created_at (created_at)
);

-- Task comments table (optional feature)
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`task_comments` (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    comment TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES `Yotei-Plus`.tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    INDEX idx_task_id (task_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);

-- Task attachments table (optional feature)
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`task_attachments` (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    uploaded_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES `Yotei-Plus`.tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (uploaded_by) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    INDEX idx_task_id (task_id)
);

-- User roles table (for more complex role management)
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`user_roles` (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    role_name VARCHAR(50) NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    granted_by VARCHAR(36) NULL,
    FOREIGN KEY (user_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES `Yotei-Plus`.users(id) ON DELETE SET NULL,
    UNIQUE KEY unique_user_role (user_id, role_name),
    INDEX idx_user_id (user_id),
    INDEX idx_role_name (role_name)
);

-- Insert sample data
INSERT INTO `Yotei-Plus`.users (id, email, username, password, role, email_verified) VALUES 
('550e8400-e29b-41d4-a716-446655440000', 'admin@example.com', 'admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', TRUE),
('550e8400-e29b-41d4-a716-446655440001', 'user@example.com', 'testuser', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user', TRUE)
ON DUPLICATE KEY UPDATE email = VALUES(email);

-- Insert sample tasks
INSERT INTO `Yotei-Plus`.tasks (id, title, description, status, priority, created_by) VALUES 
('660e8400-e29b-41d4-a716-446655440000', 'プロジェクト設計', 'アプリケーションアーキテクチャの設計', 'IN_PROGRESS', 'HIGH', '550e8400-e29b-41d4-a716-446655440000'),
('660e8400-e29b-41d4-a716-446655440001', 'データベース設計', 'ERD作成とテーブル設計', 'TODO', 'MEDIUM', '550e8400-e29b-41d4-a716-446655440000'),
('660e8400-e29b-41d4-a716-446655440002', 'API実装', 'REST API エンドポイントの実装', 'TODO', 'HIGH', '550e8400-e29b-41d4-a716-446655440001')
ON DUPLICATE KEY UPDATE title = VALUES(title);

-- Insert sample notifications
INSERT INTO `Yotei-Plus`.notifications (id, user_id, title, message, type, status) VALUES 
('770e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', 'タスクが割り当てられました', '新しいタスク「API実装」が割り当てられました', 'TASK_ASSIGNED', 'SENT'),
('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'システムメンテナンス', 'システムメンテナンスが予定されています', 'SYSTEM_NOTICE', 'PENDING')
ON DUPLICATE KEY UPDATE title = VALUES(title);

-- Note: Stored procedures removed as they may not work properly in init scripts
-- You can add them later through your application or separate migration scripts

-- Social module tables
-- Friendships table for friend relationships
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`friendships` (
    id VARCHAR(36) PRIMARY KEY,
    requester_id VARCHAR(36) NOT NULL,
    addressee_id VARCHAR(36) NOT NULL,
    status ENUM('PENDING', 'ACCEPTED', 'BLOCKED') DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP NULL,
    blocked_at TIMESTAMP NULL,
    FOREIGN KEY (requester_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    FOREIGN KEY (addressee_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_friendship (requester_id, addressee_id),
    INDEX idx_requester_id (requester_id),
    INDEX idx_addressee_id (addressee_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- Invitations table for invitation system
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`invitations` (
    id VARCHAR(36) PRIMARY KEY,
    type ENUM('FRIEND', 'GROUP') NOT NULL,
    method ENUM('IN_APP', 'CODE', 'URL') NOT NULL,
    status ENUM('PENDING', 'ACCEPTED', 'DECLINED', 'EXPIRED', 'CANCELED') DEFAULT 'PENDING',
    inviter_id VARCHAR(36) NOT NULL,
    invitee_id VARCHAR(36) NULL,
    invitee_email VARCHAR(255) NULL,
    invitee_username VARCHAR(255) NULL,
    invitee_phone VARCHAR(20) NULL,
    target_id VARCHAR(36) NULL, -- group_id for group invitations
    code VARCHAR(255) NULL,
    url TEXT NULL,
    message TEXT NULL,
    metadata JSON NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP NULL,
    FOREIGN KEY (inviter_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    FOREIGN KEY (invitee_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES `Yotei-Plus`.groups(id) ON DELETE CASCADE,
    UNIQUE KEY unique_code (code),
    INDEX idx_inviter_id (inviter_id),
    INDEX idx_invitee_id (invitee_id),
    INDEX idx_code (code),
    INDEX idx_status (status),
    INDEX idx_type (type),
    INDEX idx_expires_at (expires_at),
    INDEX idx_created_at (created_at)
);

-- Groups table
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`groups` (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    type ENUM('PROJECT', 'SCHEDULE') NOT NULL,
    owner_id VARCHAR(36) NOT NULL,
    member_count INT DEFAULT 1,
    is_public BOOLEAN DEFAULT FALSE,
    allow_member_invite BOOLEAN DEFAULT TRUE,
    require_approval BOOLEAN DEFAULT TRUE,
    enable_notifications BOOLEAN DEFAULT TRUE,
    -- Schedule group settings
    default_privacy_level ENUM('NONE', 'BUSY', 'TITLE', 'DETAILS') NULL,
    allow_schedule_details BOOLEAN NULL,
    -- Project group settings
    enable_gantt_chart BOOLEAN NULL,
    enable_task_dependency BOOLEAN NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    version INT DEFAULT 1,
    FOREIGN KEY (owner_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    INDEX idx_owner_id (owner_id),
    INDEX idx_type (type),
    INDEX idx_is_public (is_public),
    INDEX idx_created_at (created_at),
    FULLTEXT idx_search (name, description)
);

-- Group members table
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`group_members` (
    id VARCHAR(36) PRIMARY KEY,
    group_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    role ENUM('OWNER', 'ADMIN', 'MEMBER') DEFAULT 'MEMBER',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES `Yotei-Plus`.groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES `Yotei-Plus`.users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_group_member (group_id, user_id),
    INDEX idx_group_id (group_id),
    INDEX idx_user_id (user_id),
    INDEX idx_role (role),
    INDEX idx_joined_at (joined_at)
);

-- Group tasks table (extending tasks with group context)
CREATE TABLE IF NOT EXISTS `Yotei-Plus`.`group_tasks` (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    group_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES `Yotei-Plus`.tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES `Yotei-Plus`.groups(id) ON DELETE CASCADE,
    UNIQUE KEY unique_group_task (task_id, group_id),
    INDEX idx_task_id (task_id),
    INDEX idx_group_id (group_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_tasks_compound ON `Yotei-Plus`.tasks (status, assignee_id, due_date);
CREATE INDEX IF NOT EXISTS idx_notifications_compound ON `Yotei-Plus`.notifications (user_id, status, created_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_compound ON `Yotei-Plus`.refresh_tokens (user_id, expires_at, revoked_at);
CREATE INDEX IF NOT EXISTS idx_friendships_compound ON `Yotei-Plus`.friendships (requester_id, addressee_id, status);
CREATE INDEX IF NOT EXISTS idx_group_members_compound ON `Yotei-Plus`.group_members (group_id, user_id, role);
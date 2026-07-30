-- 登录授权服务 - 初始化数据库
-- 使用: mysql -u root -p < migrations/001_init.sql

CREATE DATABASE IF NOT EXISTS login DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE login;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username       VARCHAR(64)  NOT NULL UNIQUE,
    email          VARCHAR(128) NOT NULL UNIQUE,
    phone          VARCHAR(20)  NULL UNIQUE,
    password_hash  VARCHAR(256) NOT NULL,
    nickname       VARCHAR(64)  NULL,
    avatar_url     VARCHAR(512) NULL,
    status         TINYINT      NOT NULL DEFAULT 1 COMMENT '1:正常 0:禁用 -1:删除',
    email_verified TINYINT      NOT NULL DEFAULT 0,
    phone_verified TINYINT      NOT NULL DEFAULT 0,
    last_login_at  DATETIME     NULL,
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL UNIQUE,
    description VARCHAR(256) NULL,
    is_system   TINYINT      NOT NULL DEFAULT 0 COMMENT '系统内置角色不可删除',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code        VARCHAR(128) NOT NULL UNIQUE COMMENT '权限编码: user:create',
    name        VARCHAR(128) NOT NULL COMMENT '权限名称',
    module      VARCHAR(64)  NOT NULL COMMENT '所属模块',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_module (module)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 用户-角色关联
CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 角色-权限关联
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 初始化系统角色
INSERT INTO roles (name, description, is_system) VALUES
('admin', '系统管理员', 1),
('user', '普通用户', 1)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- 初始化管理员账号 (密码: admin123, 需自行修改)
INSERT INTO users (username, email, password_hash, nickname, status) VALUES
('admin', 'admin@example.com', '$2a$12$LJ3m4ys3Lk0TSwHnbfOMiOXPm1Qlq5JHgCqF9Kx5Yxh8Hnq1xGv7y', '管理员', 1)
ON DUPLICATE KEY UPDATE username = VALUES(username);

-- 分配管理员角色
INSERT IGNORE INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r WHERE u.username = 'admin' AND r.name = 'admin';

# 🔐 登录授权服务 - 构思文档 (Conceive)

> 版本: v0.1 | 日期: 2026-07-30

---

## 📋 目录

1. [项目背景与目标](#1-项目背景与目标)
2. [范围界定](#2-范围界定)
3. [技术选型](#3-技术选型)
4. [核心功能模块](#4-核心功能模块)
5. [数据库设计](#5-数据库设计)
6. [API 接口设计](#6-api-接口设计)
7. [认证流程](#7-认证流程)
8. [安全考虑](#8-安全考虑)
9. [Redis 存储设计](#9-redis-存储设计)
10. [项目结构](#10-项目结构)
11. [未来拓展](#11-未来拓展)

---

## 1. 项目背景与目标

### 1.1 为什么要做？
- **统一身份认证**：为多个子系统提供统一的登录/注册/鉴权入口
- **安全隔离**：将认证逻辑从业务代码中抽离，降低安全风险
- **可复用性**：构建标准化授权服务，后续项目可直接复用

### 1.2 核心目标
| 目标 | 描述 |
|------|------|
| 用户注册 | 支持邮箱/手机号注册 |
| 用户登录 | 支持多种登录方式 |
| Token 管理 | 签发、刷新、吊销 Access Token / Refresh Token |
| 权限控制 | 基于 RBAC 的细粒度权限管理 |
| 安全防护 | 防暴力破解、密码加密等 |

---

## 2. 范围界定

> 🎯 **核心原则**：
> - 本服务负责**颁发身份凭证** + **管理 RefreshToken 黑名单**
> - **AccessToken（15min）**：资源服务**本地验签即可**，无需查黑名单，短有效期保证安全
> - **RefreshToken（7天）**：必须可吊销，登出/刷新/改密时写入 **Redis 黑名单**
> - 黑名单 **只存 RefreshToken 的 jti**，不存 AccessToken

### 2.1 ✅ 范围内
| 职责 | 说明 |
|------|------|
| 用户注册 | 创建账户，存储凭据 |
| 用户登录 | 验证身份，颁发 Access Token + Refresh Token |
| Token 刷新 | 用 Refresh Token 换取新的 Token 对（旧 Refresh 加入黑名单） |
| Token 吊销 | 退出登录、修改密码、设备登出时吊销 RefreshToken |
| 黑名单共享 | 通过 Redis 共享黑名单，资源服务可查询 |
| 密码管理 | 修改密码、忘记/重置密码 |
| 设备管理 | 查看在线设备、踢指定设备下线 |
| 用户信息管理 | 查看/修改个人资料 |
| RBAC 权限管理 | 管理后台角色与权限的 CRUD |

### 2.2 ❌ 范围外
| 非职责 | 说明 | 由谁负责 |
|--------|------|----------|
| Token 验签 | JWT 签名验证是本地计算 | 各资源服务自行验签 |
| OAuth 2.0 第三方登录 | GitHub/Google/微信 | 中期规划 |
| 前端 Token 存储 | 前端如何存 Token、自动续期 | 前端团队 |
| 用户头像存储 | 文件/图片上传 | 独立的对象存储 |

### 2.3 服务架构图
```
                      ┌──────────────────────┐
                      │    API Gateway        │
                      └──────┬───────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
      ┌────────────┐ ┌────────────┐ ┌────────────┐
      │ 业务服务 A  │ │ 业务服务 B  │ │ 🔐 本服务   │
      │ (自验签Access)│ │ (自验签Access)│ │ 登录授权   │
      └────────────┘ └────────────┘ └──────┬─────┘
                                           │
        ┌──────────────────────────────────┘
        ▼
┌───────────────────────────────────────────────┐
│              Redis                             │
│                                                │
│  [黑名单]  blk:{refresh_jti} → revoked_at     │
│            TTL = RefreshToken 剩余有效期       │
│                                                │
│  [会话]    sessions:{user_id}                  │
│            Hash: {device_id → refresh_jti}     │
│                                                │
│  [限流]    rate_limit:{ip}:login → count      │
└───────────────────────────────────────────────┘

🔑 资源服务校验逻辑：
  1. 提取 Authorization: Bearer <access_token>
  2. 用共享公钥本地验签（RS256，无需调 Auth 服务）
  3. 检查 exp 字段 → 是否过期
  4. ✅ 通过 → 放行（AccessToken 短有效，不查黑名单）
  
  ※ 极严安全场景下，资源服务也可额外查黑名单
```

---

## 3. 技术选型

### 3.1 后端
| 方案 | 推荐 |
|------|------|
| **Go + Gin** | ⭐ 强烈推荐 |
| Node.js + NestJS | ✅ 可选 |
| Java + Spring Boot | ✅ 可选 |

### 3.2 数据库
| 组件 | 推荐 | 用途 |
|------|------|------|
| MySQL / PostgreSQL | ✅ | 用户、角色、权限 |
| Redis | ✅ | 黑名单、会话、限流 |

### 3.3 认证协议
| 协议 | 用途 |
|------|------|
| JWT (RS256) | 非对称签名，资源服务只需公钥验签 |

---

## 4. 核心功能模块

### 4.1 注册
- 邮箱注册：验证邮箱唯一性
- 密码强度：8-32位，含大小写+数字+特殊字符

### 4.2 登录
- **密码登录**：用户名/邮箱 + 密码
- **验证码登录**：手机号 + 短信验证码
- **防暴力破解**：同 IP 5次失败锁定15分钟

### 4.3 Token 管理
| 操作 | AccessToken | RefreshToken |
|------|-------------|--------------|
| 登录颁发 | ✅ 15min 有效 | ✅ 7天有效 |
| 刷新 | 生成新 AT | 生成新 RT，旧 RT 加入黑名单 |
| 登出 | 不管，让它自然过期 | 加入黑名单 |
| 修改密码 | 不管（当前设备） | 旧 RT 加入黑名单 |
| 全部登出 | 不管 | 所有 RT 加入黑名单 |

### 4.4 设备管理
- **查看设备列表**：列出用户活跃的 RefreshToken 会话
- **指定设备登出**：吊销该设备的 RefreshToken
- **全部登出**：吊销用户所有 RefreshToken

### 4.5 密码管理
- **修改密码**：旧密码 → 新密码 → 吊销其他设备 RT
- **忘记密码**：验证码 → 新密码 → 吊销所有 RT
- **重置密码**（管理员）：强制重置

### 4.6 RBAC 权限
- **用户 → 角色 → 权限** 三层模型

---

## 5. 数据库设计

### 5.1 用户表 users
```sql
CREATE TABLE users (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username       VARCHAR(64)  NOT NULL UNIQUE,
    email          VARCHAR(128) NOT NULL UNIQUE,
    phone          VARCHAR(20)  NULL UNIQUE,
    password_hash  VARCHAR(256) NOT NULL,        -- bcrypt
    nickname       VARCHAR(64)  NULL,
    status         TINYINT      NOT NULL DEFAULT 1, -- 1正常 0禁用 -1删除
    email_verified TINYINT      NOT NULL DEFAULT 0,
    phone_verified TINYINT      NOT NULL DEFAULT 0,
    last_login_at  DATETIME     NULL,
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 5.2 角色表 roles
```sql
CREATE TABLE roles (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL UNIQUE,
    description VARCHAR(256) NULL,
    is_system   TINYINT      NOT NULL DEFAULT 0, -- 系统内置不可删
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 5.3 权限表 permissions
```sql
CREATE TABLE permissions (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code        VARCHAR(128) NOT NULL UNIQUE,  -- user:create, article:edit
    name        VARCHAR(128) NOT NULL,
    module      VARCHAR(64)  NOT NULL,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 5.4 关联表
```sql
-- 用户-角色
CREATE TABLE user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- 角色-权限
CREATE TABLE role_permissions (
    role_id       BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);
```

---

## 6. API 接口设计

### 6.1 认证 Auth
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/v1/auth/register | 注册 | ❌ |
| POST | /api/v1/auth/login | 密码登录 | ❌ |
| POST | /api/v1/auth/login/code | 验证码登录 | ❌ |
| POST | /api/v1/auth/refresh | 刷新 Token | ✅ Refresh |
| POST | /api/v1/auth/logout | 登出（吊销当前 RT） | ✅ Access |
| POST | /api/v1/auth/logout/device/:deviceId | 指定设备登出 | ✅ Access |
| POST | /api/v1/auth/logout/all | 全部设备登出 | ✅ Access |
| POST | /api/v1/auth/forgot-password | 忘记密码 | ❌ |
| POST | /api/v1/auth/reset-password | 重置密码 | ❌ |

### 6.2 用户 User
| 方法 | 路径 | 说明 | 认证 | 权限 |
|------|------|------|------|------|
| GET | /api/v1/users/me | 个人信息 | ✅ | - |
| PUT | /api/v1/users/me | 修改资料 | ✅ | - |
| PUT | /api/v1/users/me/password | 修改密码 | ✅ | - |
| GET | /api/v1/users/me/devices | 设备列表 | ✅ | - |
| GET | /api/v1/users | 用户列表 | ✅ | admin |
| PUT | /api/v1/users/:id/status | 启用/禁用 | ✅ | admin |

### 6.3 角色 Role / 权限 Permission
| 方法 | 路径 | 说明 | 认证 | 权限 |
|------|------|------|------|------|
| GET/POST | /api/v1/roles | 列表/创建 | ✅ | admin |
| PUT/DELETE | /api/v1/roles/:id | 修改/删除 | ✅ | admin |
| PUT | /api/v1/roles/:id/permissions | 分配权限 | ✅ | admin |
| GET/POST | /api/v1/permissions | 列表/创建 | ✅ | admin |

### 6.4 响应格式
```json
// 成功
{ "code": 0, "message": "success", "data": { ... } }

// 登录返回
{ "code": 0, "data": {
    "access_token": "eyJ...",
    "refresh_token": "dGhp...",
    "expires_in": 900,
    "token_type": "Bearer",
    "user": { "id": 1, "username": "admin", "roles": ["admin"] }
}}

// 错误
{ "code": 40101, "message": "用户名或密码错误", "data": null }
```

### 6.5 错误码
| 码 | 说明 | 码 | 说明 |
|----|------|----|------|
| 0 | 成功 | 40102 | Token 过期 |
| 40000 | 参数错误 | 40103 | Token 已吊销 |
| 40100 | 未认证 | 40300 | 无权限 |
| 40101 | 账号或密码错 | 42900 | 请求频繁 |


---

## 7. 认证流程

### 7.1 登录
```
Client              Auth Service              DB              Redis
  │                     │                      │                │
  │ POST /login         │                      │                │
  │────────────────────►│                      │                │
  │                     │ 查IP是否锁定          │                │
  │                     │──────────────────────────────────────►│
  │                     │◄──────────────────────────────────────│
  │                     │ 查用户 + 比对密码     │                │
  │                     │─────────────────────►│                │
  │                     │◄─────────────────────│                │
  │                     │ 记录登录失败次数      │                │
  │                     │──────────────────────────────────────►│
  │                     │ 生成 AT(15min)+RT(7d)│                │
  │                     │ 存 RT 到 sessions    │                │
  │                     │──────────────────────────────────────►│
  │◄── 200 + AT + RT ──│                      │                │
```

### 7.2 刷新 Token
```
Client              Auth Service                            Redis
  │                     │                                      │
  │ POST /refresh       │                                      │
  │ refresh_token=xxx   │                                      │
  │────────────────────►│                                      │
  │                     │ 验签 RefreshToken                    │
  │                     │ 查黑名单 EXISTS blk:{jti}            │
  │                     │─────────────────────────────────────►│
  │                     │◄─────────────────────────────────────│
  │                     │ 旧 RT 加入黑名单                     │
  │                     │─────────────────────────────────────►│
  │                     │ 生成新 AT + 新 RT                    │
  │                     │ 更新 sessions                        │
  │◄── 200 + 新 AT+RT ─│                                      │
```

### 7.3 登出
```
Client              Auth Service                            Redis
  │                     │                                      │
  │ POST /logout        │                                      │
  │ Authorization: AT   │                                      │
  │ refresh_token=xxx   │                                      │
  │────────────────────►│                                      │
  │                     │ 验签 AT（获取用户身份）               │
  │                     │ 将 RefreshToken jti 加入黑名单       │
  │                     │─────────────────────────────────────►│
  │                     │ 删除 sessions 记录                   │
  │                     │─────────────────────────────────────►│
  │◄── 200 ────────────│                                      │
  │                     │                                      │
  │ 前端清除本地 AT+RT  │                                      │
```

### 7.4 修改密码
```
Client              Auth Service                            Redis
  │                     │                                      │
  │ PUT /me/password    │                                      │
  │────────────────────►│                                      │
  │                     │ 验证旧密码 → 更新新密码              │
  │                     │ 吊销除当前设备外的所有 RT            │
  │                     │─────────────────────────────────────►│
  │                     │ 生成新 AT + 新 RT (当前设备)         │
  │◄── 200 + 新 AT+RT ─│                                      │
```

---

## 8. 安全考虑

### 8.1 密码安全
- **bcrypt**（cost=10+），不存明文
- 密码强度：8-32位，大小写+数字+特殊字符
- 弱密码黑名单

### 8.2 Token 安全
- **RS256 非对称签名**：Auth 服务私钥签名，资源服务公钥验签
- AccessToken：15min，短有效减少攻击窗口
- RefreshToken：7天，可吊销
- **Token Rotation**：每次刷新，旧 RefreshToken 立即失效

### 8.3 防攻击
| 类型 | 防护 |
|------|------|
| 暴力破解 | IP限流 + 账户锁定（5次/15min） |
| CSRF | SameSite Cookie |
| XSS | HttpOnly Cookie / 前端内存 |
| SQL注入 | 参数化查询 |
| 用户枚举 | 登录失败统一提示 |

---

## 9. Redis 存储设计

> ⚠️ **关键约定**：黑名单**只存 RefreshToken**，不存 AccessToken

### 9.1 黑名单（吊销的 RefreshToken）
```
Key:   blk:{refresh_jti}
Value: revoked_at (Unix 时间戳)
TTL:   RefreshToken 剩余有效期（最长 7d）
功能:  EXISTS 判断是否被吊销

使用场景：
  - 用户登出    → 当前 RT 加入黑名单
  - 刷新 Token  → 旧 RT 加入黑名单
  - 设备登出    → 指定设备的 RT 加入黑名单
  - 全部登出    → 该用户所有 RT 加入黑名单
  - 修改密码    → 其他设备 RT 加入黑名单
  - 忘记密码    → 所有 RT 加入黑名单
```

### 9.2 活跃会话（设备管理用）
```
Key:   sessions:{user_id}
Type:  Hash
Field: {device_id}（设备唯一标识）
Value: {"refresh_jti":"xxx","user_agent":"...","ip":"...","created_at":时间戳}
TTL:   无（由 RT 有效期控制）

功能：
  - 查看用户登录设备
  - 设备登出时，找到 refresh_jti 加入黑名单
```


### 9.3 限流计数器
```
Key:   rate_limit:login:{ip}
Value: 尝试次数（INCR）
TTL:   15 分钟
功能:  防 IP 级别暴力破解

Key:   rate_limit:login:{user_id}
Value: 尝试次数
TTL:   15 分钟
功能:  防针对特定账号的暴力破解
```

### 9.4 场景示例

**登录成功 → Redis：**
```
# 存会话
HSET sessions:{user_id} {device_id} '{"refresh_jti":"abc123","user_agent":"Chrome/120","ip":"1.2.3.4","created_at":1690000000}'
```

**登出 → Redis：**
```
# 吊销 RefreshToken（只吊销 RT）
SET blk:abc123 1690000100 EX 604800
# 删除会话
HDEL sessions:{user_id} {device_id}
# 前端自行清除本地 AT
```

**资源服务验签（无需查 Redis）：**
```
1. 提取 AT
2. 公钥验 RS256 签名
3. 检查 exp
4. 通过 → 放行 ✅
```

---

## 10. 项目结构

```
login/
├── conceive.md
├── cmd/server/main.go
├── internal/
│   ├── config/config.go
│   ├── handler/
│   │   ├── auth.go        # 登录/注册/登出/刷新
│   │   ├── user.go        # 用户管理
│   │   ├── role.go        # 角色管理
│   │   └── permission.go  # 权限管理
│   ├── service/
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── role.go
│   │   └── permission.go
│   ├── repository/
│   ├── model/
│   ├── dto/
│   ├── middleware/         # cors, logger, recovery
│   ├── pkg/
│   │   ├── jwt/jwt.go     # JWT 工具
│   │   ├── hash/bcrypt.go # 密码哈希
│   │   └── response.go    # 统一响应
│   └── router/router.go
├── pkg/database/mysql.go
├── pkg/cache/redis.go
├── migrations/
├── config/
├── Makefile
└── README.md
```

---

## 11. 未来拓展

### 11.1 MVP 版本（第一期）
- [x] 用户注册（邮箱）
- [x] 用户登录（密码 + 验证码）
- [x] 密码管理（修改密码、忘记/重置密码）
- [x] 设备管理（查看设备、指定登出、全部登出）
- [x] JWT 签发 + Token Rotation（RS256）
- [x] RefreshToken 黑名单（Redis）
- [x] 防暴力破解（IP + 账户限流）
- [x] 用户信息管理（查看/修改）
- [x] RBAC（角色与权限 CRUD）
- [x] 统一响应格式 & 错误码

### 11.2 第二期
- [ ] OAuth 2.0 第三方登录
- [ ] 短信/邮件验证码
- [ ] 双因素认证 (2FA)
- [ ] 登录日志 & 审计

### 11.3 第三期
- [ ] 单点登录 (SSO)
- [ ] OIDC 协议
- [ ] WebAuthn 无密码登录

---

## 🤔 待决策

1. **技术栈**：Go + Gin 还是其他？
2. **数据库**：MySQL 还是 PostgreSQL？
3. **ORM**：GORM 还是 sqlx？
4. **Token 签名**：RS256（非对称）还是 HS256（对称）？
5. **日志库**：Zap 还是 Logrus？
6. **首次迭代**：先做哪些接口？

---

> 📝 **下一步**：确认后搭建项目骨架，从 `main.go` + 数据库迁移开始逐步实现！

| OAuth 2.0 | 第三方登录（后续迭代） |

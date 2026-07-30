# 🔐 Login Auth Service

> 基于 Go + Gin 的登录授权服务，提供统一身份认证与 Token 管理。

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-1.12-0097d3)](https://gin-gonic.com/)
[![License](https://img.shields.io/badge/License-MIT-green)]()

## 📌 一句话描述

**登录授权服务** — 用户注册登录、JWT 签发刷新吊销、RBAC 权限管理，资源服务自行验签，Auth 服务只管理黑名单。

---

## 🏗 架构概要

```
Client → Auth Service → 颁发 AT(15min) + RT(7d)
                      → RT 黑名单写入 Redis (blk:{jti})
                      → 用户 RT 集合写入 Redis (user_tokens:{uid})

Client → Resource Service → 公钥本地验签 AT
                          → 无需查黑名单（AT 短有效）
```

**核心原则：**
- **AccessToken（15min）**：资源服务自行验签，不落黑名单
- **RefreshToken（7天）**：可吊销，登出/刷新/改密时入黑名单
- **黑名单只存 RT 的 jti**，不存 AT

---

## ✨ 功能列表

| 接口 | 说明 | 认证 |
|------|------|------|
| `POST /api/v1/auth/register` | 注册（自动登录返回 Token） | ❌ |
| `POST /api/v1/auth/login` | 密码登录 | ❌ |
| `POST /api/v1/auth/refresh` | 刷新 Token | ✅ RT |
| `POST /api/v1/auth/logout` | 登出（吊销当前 RT） | ✅ AT |
| `POST /api/v1/auth/logout/device/:id` | 指定设备登出 | ✅ AT |
| `POST /api/v1/auth/logout/all` | 全部设备登出 | ✅ AT |
| `PUT /api/v1/auth/password` | 修改密码 | ✅ AT |
| `GET /api/v1/users/me` | 个人信息 | ✅ AT |
| `GET /api/v1/users/me/devices` | 设备列表 | ✅ AT |
| `GET /api/v1/users` | 用户列表（管理） | ✅ AT+Admin |
| `GET/POST /api/v1/roles` | 角色管理 | ✅ AT+Admin |
| `GET/POST /api/v1/permissions` | 权限管理 | ✅ AT+Admin |

---

## 🚀 快速开始

### 前置条件

- Go 1.25+
- MySQL 8.0+
- Redis 7.0+

### 1. 克隆 & 配置

```bash
git clone https://github.com/your/login.git
cd login

# 修改数据库/Redis 配置
vim config/config.yaml
```

### 2. 初始化数据库

```bash
mysql -u root -p < migrations/001_init.sql
```

### 3. 生成 JWT 密钥

```bash
# 任意方式生成 RSA 2048 密钥对，放 config/ 目录即可
# 例如用 Go 自带的 crypto 生成（项目初次启动自动生成也可以）

# 或者用 OpenSSL（如果有）：
openssl genrsa -out config/rsa_private.pem 2048
openssl rsa -in config/rsa_private.pem -pubout -out config/rsa_public.pem
```

> 注：也可以不手动生成，项目启动时若密钥文件不存在，会自动创建（后续迭代）。

### 4. 启动

```bash
make run
# 服务默认监听 :8080
```

### 5. 测试

```bash
# 注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@test.com","password":"Test1234!"}'

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

---

## ⚙️ 配置说明

```yaml
server:
  port: 8080
  mode: debug

database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: root
  dbname: login

jwt:
  access_token_ttl: 15m
  refresh_token_ttl: 168h  # 7天
```

---

## 🗄 Redis 设计

```
blk:{jti}              → String, TTL=RT剩余有效期  (黑名单)
user_tokens:{user_id}  → Set, 无TTL               (用户活跃RT集合)
rate_limit:{key}       → String, TTL=15min        (限流计数器)
```

---

## 📁 项目结构

```
login/
├── cmd/server/main.go       # 入口
├── config/                  # 配置 + RSA密钥
├── internal/
│   ├── config/              # 配置加载
│   ├── handler/             # HTTP处理器
│   ├── service/             # 业务逻辑
│   ├── repository/          # 数据访问
│   ├── model/               # 数据模型
│   ├── dto/                 # 请求/响应
│   ├── middleware/          # CORS/日志/恢复/JWT鉴权
│   ├── pkg/jwt/             # JWT工具(RS256)
│   ├── pkg/hash/            # bcrypt
│   └── router/              # 路由注册
├── pkg/database/            # MySQL连接
├── pkg/cache/               # Redis客户端
└── migrations/              # SQL迁移
```

---

## 🧰 常用命令

```bash
make build    # 编译
make run      # 运行
make gen-key  # 生成RSA密钥
make db-init  # 初始化数据库
make tidy     # 整理依赖
make fmt      # 格式化代码
```

---

## 📄 架构设计文档

详细的设计思路和决策过程见 [`conceive_arch.md`](./conceive_arch.md)。

---

## 🔜 Roadmap

- [ ] OAuth 2.0 第三方登录
- [ ] 短信/邮件验证码
- [ ] 双因素认证 (2FA)
- [ ] 登录审计日志
- [ ] 单点登录 (SSO)

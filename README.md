# 🔐 Login Auth Service

> 基于 Go + Gin 的登录授权的单体服务，提供统一身份认证与 Token 管理。

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
make gen-key
```


### 4. 启动后端

```bash
make run
# 服务监听 :8080
```

### 5. 启动前端

项目包含 Vite + React 前端，两种方式：

**开发模式**（热重载，推荐）：
```bash
cd frontend
npm install        # 首次需要
npm run dev        # 打开 http://localhost:5173
# Vite 自动代理 /api 到后端 :8080
```

**生产模式**（编译后由 Gin 托管）：
```bash
cd frontend
npm run build
cd ..
make run           # 打开 http://localhost:8080
```

### 6. 测试

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
├── cmd/server/main.go       # 程序入口，启动 HTTP 服务
├── config/                  # 配置文件 + RSA 密钥对
├── internal/
│   ├── config/config.go     # 配置结构体定义与加载 yaml
│   ├── handler/             # HTTP 请求处理器（路由回调）
│   ├── service/             # 核心业务逻辑层
│   ├── repository/          # 数据访问层（GORM 操作 MySQL）
│   ├── model/               # GORM 数据模型（表映射）
│   ├── dto/                 # 请求/响应数据结构体
│   ├── middleware/          # Gin 中间件（CORS/日志/恢复/JWT鉴权）
│   ├── pkg/jwt/             # JWT 工具（RS256 签名与验签）
│   ├── pkg/hash/            # bcrypt 密码加密
│   └── router/              # 路由注册，组装中间件与处理器
├── pkg/database/            # MySQL 连接初始化
├── pkg/cache/               # Redis 客户端（黑名单/用户Token集/限流）
└── migrations/              # MySQL 初始化建表脚本
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

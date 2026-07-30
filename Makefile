.PHONY: build run test clean gen-key

APP_NAME = login-server
BUILD_DIR = build

# 构建
build:
	@echo "Building $(APP_NAME)..."
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

# 运行
run:
	go run ./cmd/server

# 开发模式运行（文件热重载，需安装 air）
dev:
	air

# 测试
test:
	go test ./... -v

# 生成 RSA 密钥对（JWT 签名用）
gen-key:
	@echo "Generating RSA key pair..."
	openssl genrsa -out config/rsa_private.pem 2048
	openssl rsa -in config/rsa_private.pem -pubout -out config/rsa_public.pem
	@echo "Done! RSA keys generated in config/"

# 初始化数据库（需先配置 MySQL）
db-init:
	mysql -u root -p < migrations/001_init.sql

# 代码格式化
fmt:
	go fmt ./...

# 依赖整理
tidy:
	go mod tidy

# 静态分析
lint:
	golangci-lint run ./...

# 清理
clean:
	rm -rf $(BUILD_DIR)
	rm -rf logs/

# 全部
all: tidy fmt build

package main

import (
	"context"
	"fmt"
	"login/internal/config"
	"login/internal/handler"
	"login/internal/pkg/jwt"
	"login/internal/repository"
	"login/internal/router"
	"login/internal/service"
	"login/pkg/cache"
	"login/pkg/database"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 初始化日志
	log, err := initLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("初始化日志失败: %v", err))
	}
	defer log.Sync()

	// 初始化数据库
	db, err := database.NewMySQL(&cfg.Database, log)
	if err != nil {
		log.Fatal("数据库连接失败", zap.Error(err))
	}

	// 初始化Redis
	redisClient, err := cache.NewRedis(&cfg.Redis, log)
	if err != nil {
		log.Fatal("Redis连接失败", zap.Error(err))
	}
	defer redisClient.Close()

	// 初始化JWT
	jwtManager, err := jwt.New(
		cfg.JWT.PrivateKeyPath,
		cfg.JWT.PublicKeyPath,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)
	if err != nil {
		log.Fatal("初始化JWT失败", zap.Error(err))
	}

	// 初始化仓储层
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)

	// 初始化服务层
	authService := service.NewAuthService(userRepo, jwtManager, redisClient, log)
	userService := service.NewUserService(userRepo, log)
	roleService := service.NewRoleService(roleRepo, log)
	permService := service.NewPermissionService(permRepo, log)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService, log)
	userHandler := handler.NewUserHandler(userService, authService, log)
	roleHandler := handler.NewRoleHandler(roleService, log)
	permHandler := handler.NewPermissionHandler(permService, log)

	// 设置路由
	r := router.Setup(authHandler, userHandler, roleHandler, permHandler, jwtManager, log)

	// 启动服务
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info("服务启动", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务关闭异常", zap.Error(err))
	}
	log.Info("服务已关闭")
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var log *zap.Logger
	var err error

	switch cfg.Log.Level {
	case "debug":
		log, err = zap.NewDevelopment()
	default:
		log, err = zap.NewProduction()
	}
	return log, err
}

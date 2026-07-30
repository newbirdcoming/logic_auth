package router

import (
	"login/internal/handler"
	"login/internal/middleware"
	"login/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Setup(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
	permHandler *handler.PermissionHandler,
	jwtManager *jwt.JWTManager,
	log *zap.Logger,
) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))

	// 前端静态文件
	r.Static("/assets", "./frontend/dist/assets")
	r.GET("/", func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// 需要登录的接口
	authed := api.Group("")
	authed.Use(middleware.AuthRequired(jwtManager, log))
	{
		// 认证相关
		authed.POST("/auth/logout", authHandler.Logout)
		authed.POST("/auth/logout/device/:deviceId", authHandler.LogoutDevice)
		authed.POST("/auth/logout/all", authHandler.LogoutAll)
		authed.PUT("/auth/password", authHandler.ChangePassword)

		// 用户相关
		authed.GET("/users/me", userHandler.GetProfile)
		authed.PUT("/users/me", userHandler.UpdateProfile)
		authed.GET("/users/me/devices", userHandler.GetDevices)

		// 管理员接口
		admin := authed.Group("/admin")
		{
			admin.GET("/users", userHandler.List)
			admin.PUT("/users/:id/status", userHandler.UpdateStatus)

			admin.GET("/roles", roleHandler.List)
			admin.POST("/roles", roleHandler.Create)
			admin.PUT("/roles/:id", roleHandler.Update)
			admin.DELETE("/roles/:id", roleHandler.Delete)
			admin.PUT("/roles/:id/permissions", roleHandler.AssignPermissions)

			admin.GET("/permissions", permHandler.List)
			admin.POST("/permissions", permHandler.Create)
		}
	}

	return r
}

package middleware

import (
	"login/internal/pkg"
	"login/internal/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthRequired(jwtManager *jwt.JWTManager, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			pkg.Unauthorized(c, 40100, "未认证")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := jwtManager.ValidateToken(tokenStr)
		if err != nil {
			log.Warn("token验证失败", zap.String("token", tokenStr[:min(20, len(tokenStr))]))
			pkg.Unauthorized(c, 40103, "Token无效或已过期")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("device_id", claims.DeviceID)
		c.Next()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

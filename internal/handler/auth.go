package handler

import (
	"errors"
	"login/internal/dto"
	"login/internal/pkg"
	"login/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService *service.AuthService
	log         *zap.Logger
}

func NewAuthHandler(authService *service.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, log: log}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = "default"
	}
	resp, err := h.authService.Register(&req, deviceID, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			pkg.BadRequest(c, "用户名或邮箱已存在")
			return
		}
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = "default"
	}
	resp, err := h.authService.Login(&req, deviceID, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRateLimited):
			pkg.TooManyRequests(c, "登录过于频繁，请稍后再试")
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, service.ErrInvalidPassword):
			pkg.Unauthorized(c, 40101, "用户名或密码错误")
		case errors.Is(err, service.ErrUserDisabled):
			pkg.Unauthorized(c, 40100, "账号已被禁用")
		default:
			pkg.InternalError(c)
		}
		return
	}
	pkg.Success(c, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = "default"
	}
	resp, err := h.authService.Refresh(req.RefreshToken, deviceID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			pkg.Unauthorized(c, 40103, "Token无效")
			return
		}
		if errors.Is(err, service.ErrTokenRevoked) {
			pkg.Unauthorized(c, 40103, "Token已被吊销")
			return
		}
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetUint64("user_id")
	deviceID := c.GetString("device_id")
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	if err := h.authService.Logout(userID, req.RefreshToken, deviceID); err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, nil)
}

func (h *AuthHandler) LogoutDevice(c *gin.Context) {
	userID := c.GetUint64("user_id")
	deviceID := c.Param("deviceId")
	if err := h.authService.LogoutDevice(userID, deviceID); err != nil {
		pkg.BadRequest(c, "设备未找到")
		return
	}
	pkg.Success(c, nil)
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := c.GetUint64("user_id")
	deviceID := c.GetString("device_id")
	if err := h.authService.LogoutAll(userID, deviceID); err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, nil)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint64("user_id")
	deviceID := c.GetString("device_id")
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	resp, err := h.authService.ChangePassword(userID, deviceID, req.OldPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			pkg.BadRequest(c, "旧密码错误")
			return
		}
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, resp)
}

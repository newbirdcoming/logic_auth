package handler

import (
	"login/internal/dto"
	"login/internal/pkg"
	"login/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService *service.UserService
	authService *service.AuthService
	log         *zap.Logger
}

func NewUserHandler(userService *service.UserService, authService *service.AuthService, log *zap.Logger) *UserHandler {
	return &UserHandler{userService: userService, authService: authService, log: log}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint64("user_id")
	info, err := h.userService.GetProfile(userID)
	if err != nil {
		pkg.NotFound(c, "用户不存在")
		return
	}
	pkg.Success(c, info)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint64("user_id")
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	if err := h.userService.UpdateProfile(userID, &req); err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, nil)
}

func (h *UserHandler) GetDevices(c *gin.Context) {
	userID := c.GetUint64("user_id")
	devices, err := h.authService.GetDevices(userID)
	if err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, devices)
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	users, total, err := h.userService.List(page, pageSize)
	if err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, gin.H{"list": users, "total": total})
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	if err := h.userService.UpdateStatus(id, req.Status); err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, nil)
}

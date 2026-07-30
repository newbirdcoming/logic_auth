package handler

import (
	"login/internal/dto"
	"login/internal/pkg"
	"login/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RoleHandler struct {
	roleService *service.RoleService
	log         *zap.Logger
}

func NewRoleHandler(roleService *service.RoleService, log *zap.Logger) *RoleHandler {
	return &RoleHandler{roleService: roleService, log: log}
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	role, err := h.roleService.Create(&req)
	if err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, role)
}

func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.roleService.List()
	if err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, roles)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	if err := h.roleService.Update(id, &req); err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, nil)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.roleService.Delete(id); err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, nil)
}

func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	if err := h.roleService.AssignPermissions(id, &req); err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, nil)
}

type PermissionHandler struct {
	permService *service.PermissionService
	log         *zap.Logger
}

func NewPermissionHandler(permService *service.PermissionService, log *zap.Logger) *PermissionHandler {
	return &PermissionHandler{permService: permService, log: log}
}

func (h *PermissionHandler) Create(c *gin.Context) {
	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	p, err := h.permService.Create(&req)
	if err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, p)
}

func (h *PermissionHandler) List(c *gin.Context) {
	permissions, err := h.permService.List()
	if err != nil {
		pkg.InternalError(c)
		return
	}
	pkg.Success(c, permissions)
}

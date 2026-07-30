package service

import (
	"login/internal/dto"
	"login/internal/model"
	"login/internal/repository"

	"go.uber.org/zap"
)

type RoleService struct {
	roleRepo *repository.RoleRepository
	log      *zap.Logger
}

func NewRoleService(roleRepo *repository.RoleRepository, log *zap.Logger) *RoleService {
	return &RoleService{roleRepo: roleRepo, log: log}
}

func (s *RoleService) Create(req *dto.CreateRoleRequest) (*model.Role, error) {
	role := &model.Role{Name: req.Name, Description: req.Description}
	err := s.roleRepo.Create(role)
	return role, err
}

func (s *RoleService) List() ([]model.Role, error) {
	return s.roleRepo.List()
}

func (s *RoleService) Update(id uint64, req *dto.UpdateRoleRequest) error {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return err
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	return s.roleRepo.Update(role)
}

func (s *RoleService) Delete(id uint64) error {
	return s.roleRepo.Delete(id)
}

func (s *RoleService) AssignPermissions(id uint64, req *dto.AssignPermissionsRequest) error {
	return s.roleRepo.AssignPermissions(id, req.PermissionIDs)
}

// PermissionService
type PermissionService struct {
	permRepo *repository.PermissionRepository
	log      *zap.Logger
}

func NewPermissionService(permRepo *repository.PermissionRepository, log *zap.Logger) *PermissionService {
	return &PermissionService{permRepo: permRepo, log: log}
}

func (s *PermissionService) Create(req *dto.CreatePermissionRequest) (*model.Permission, error) {
	p := &model.Permission{Code: req.Code, Name: req.Name, Module: req.Module}
	err := s.permRepo.Create(p)
	return p, err
}

func (s *PermissionService) List() ([]model.Permission, error) {
	return s.permRepo.List()
}

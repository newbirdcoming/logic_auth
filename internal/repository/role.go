package repository

import (
	"login/internal/model"

	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *RoleRepository) FindByID(id uint64) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	return &role, err
}

func (r *RoleRepository) FindByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *RoleRepository) List() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

func (r *RoleRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Role{}, id).Error
}

func (r *RoleRepository) AssignPermissions(roleID uint64, permissionIDs []uint64) error {
	var role model.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}
	var permissions []model.Permission
	if err := r.db.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		return err
	}
	return r.db.Model(&role).Association("Permissions").Replace(&permissions)
}

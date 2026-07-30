package repository

import (
	"login/internal/model"

	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) Create(p *model.Permission) error {
	return r.db.Create(p).Error
}

func (r *PermissionRepository) FindByID(id uint64) (*model.Permission, error) {
	var p model.Permission
	err := r.db.First(&p, id).Error
	return &p, err
}

func (r *PermissionRepository) List() ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

func (r *PermissionRepository) ListByIDs(ids []uint64) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Where("id IN ?", ids).Find(&permissions).Error
	return permissions, err
}

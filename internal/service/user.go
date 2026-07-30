package service

import (
	"login/internal/dto"
	"login/internal/model"
	"login/internal/repository"

	"go.uber.org/zap"
)

type UserService struct {
	userRepo *repository.UserRepository
	log      *zap.Logger
}

func NewUserService(userRepo *repository.UserRepository, log *zap.Logger) *UserService {
	return &UserService{userRepo: userRepo, log: log}
}

func (s *UserService) GetProfile(id uint64) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	info := dto.ToUserInfo(user)
	return &info, nil
}

func (s *UserService) UpdateProfile(id uint64, req *dto.UpdateUserRequest) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	return s.userRepo.Update(user)
}

func (s *UserService) List(page, pageSize int) ([]model.User, int64, error) {
	return s.userRepo.List(page, pageSize)
}

func (s *UserService) UpdateStatus(id uint64, status int8) error {
	return s.userRepo.UpdateStatus(id, status)
}

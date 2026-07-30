package service

import (
	"context"
	"errors"
	"login/internal/dto"
	"login/internal/model"
	"login/internal/pkg/hash"
	"login/internal/pkg/jwt"
	"login/internal/repository"
	"login/pkg/cache"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrUserExists      = errors.New("用户已存在")
	ErrUserNotFound    = errors.New("用户不存在")
	ErrInvalidPassword = errors.New("密码错误")
	ErrUserDisabled    = errors.New("账号已被禁用")
	ErrInvalidToken    = errors.New("Token无效")
	ErrTokenRevoked    = errors.New("Token已被吊销")
	ErrRateLimited     = errors.New("请求过于频繁")
	ErrDeviceNotFound  = errors.New("设备未找到")
)

const (
	MaxLoginAttempts = 5
	RateLimitTTL     = 15 * time.Minute
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *jwt.JWTManager
	redis      *cache.RedisClient
	log        *zap.Logger
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *jwt.JWTManager, redis *cache.RedisClient, log *zap.Logger) *AuthService {
	return &AuthService{userRepo: userRepo, jwtManager: jwtManager, redis: redis, log: log}
}

func (s *AuthService) Register(req *dto.RegisterRequest, deviceID, ip, userAgent string) (*dto.LoginResponse, error) {
	exists, _ := s.userRepo.ExistsByUsername(req.Username)
	if exists {
		return nil, ErrUserExists
	}
	exists, _ = s.userRepo.ExistsByEmail(req.Email)
	if exists {
		return nil, ErrUserExists
	}

	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: passwordHash,
		Nickname:     req.Nickname,
		Status:       1,
	}
	if err := s.userRepo.Create(user); err != nil {
		s.log.Error("创建用户失败", zap.Error(err))
		return nil, err
	}

	return s.generateTokensAndSession(user, deviceID, ip, userAgent)
}

func (s *AuthService) Login(req *dto.LoginRequest, deviceID, ip, userAgent string) (*dto.LoginResponse, error) {
	ctx := context.Background()
	count, _ := s.redis.IncrementRateLimit(ctx, "login:"+ip, RateLimitTTL)
	if count > MaxLoginAttempts {
		return nil, ErrRateLimited
	}

	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if user.Status != 1 {
		return nil, ErrUserDisabled
	}
	if !hash.CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidPassword
	}

	s.redis.ResetRateLimit(ctx, "login:"+ip)
	s.redis.ResetRateLimit(ctx, "login:user:"+formatUint64(user.ID))
	s.userRepo.UpdateLastLogin(user.ID)
	return s.generateTokensAndSession(user, deviceID, ip, userAgent)
}

func (s *AuthService) Refresh(refreshTokenStr, deviceID string) (*dto.TokenResponse, error) {
	ctx := context.Background()
	claims, err := s.jwtManager.ValidateToken(refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidToken
	}
	blacklisted, _ := s.redis.IsBlacklisted(ctx, claims.ID)
	if blacklisted {
		return nil, ErrTokenRevoked
	}
	s.redis.AddToBlacklist(ctx, claims.ID, s.jwtManager.RefreshTTL())

	accessToken, expiresIn, _ := s.jwtManager.GenerateAccessToken(claims.UserID, claims.Username, claims.Roles, deviceID)
	newRefreshToken, _, newJTI, _ := s.jwtManager.GenerateRefreshToken(claims.UserID, deviceID)

	s.redis.SaveSession(ctx, claims.UserID, deviceID, &cache.SessionData{RefreshJTI: newJTI, CreatedAt: time.Now().Unix()})

	return &dto.TokenResponse{AccessToken: accessToken, RefreshToken: newRefreshToken, ExpiresIn: expiresIn, TokenType: "Bearer"}, nil
}

func (s *AuthService) Logout(userID uint64, refreshTokenStr, deviceID string) error {
	ctx := context.Background()
	claims, err := s.jwtManager.ValidateToken(refreshTokenStr)
	if err != nil {
		return ErrInvalidToken
	}
	s.redis.AddToBlacklist(ctx, claims.ID, s.jwtManager.RefreshTTL())
	s.redis.DeleteSession(ctx, userID, deviceID)
	return nil
}

func (s *AuthService) LogoutDevice(userID uint64, deviceID string) error {
	ctx := context.Background()
	session, err := s.redis.GetSession(ctx, userID, deviceID)
	if err != nil {
		return ErrDeviceNotFound
	}
	s.redis.AddToBlacklist(ctx, session.RefreshJTI, s.jwtManager.RefreshTTL())
	s.redis.DeleteSession(ctx, userID, deviceID)
	return nil
}

func (s *AuthService) LogoutAll(userID uint64, excludeDeviceID string) error {
	ctx := context.Background()
	sessions, _ := s.redis.GetAllSessions(ctx, userID)
	for did, session := range sessions {
		if did == excludeDeviceID {
			continue
		}
		s.redis.AddToBlacklist(ctx, session.RefreshJTI, s.jwtManager.RefreshTTL())
	}
	s.redis.DeleteAllSessions(ctx, userID)
	return nil
}

func (s *AuthService) GetDevices(userID uint64) ([]*dto.DeviceInfo, error) {
	ctx := context.Background()
	sessions, err := s.redis.GetAllSessions(ctx, userID)
	if err != nil {
		return nil, err
	}
	devices := make([]*dto.DeviceInfo, 0, len(sessions))
	for deviceID, session := range sessions {
		devices = append(devices, &dto.DeviceInfo{
			DeviceID:     deviceID,
			UserAgent:    session.UserAgent,
			IP:           session.IP,
			CreatedAt:    session.CreatedAt,
			LastAccessAt: session.LastAccessAt,
		})
	}
	return devices, nil
}

func (s *AuthService) ChangePassword(userID uint64, currentDeviceID string, oldPassword, newPassword string) (*dto.TokenResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if !hash.CheckPassword(oldPassword, user.PasswordHash) {
		return nil, ErrInvalidPassword
	}
	newHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return nil, err
	}
	s.userRepo.UpdatePassword(userID, newHash)

	// 吊销其他设备的RT
	ctx := context.Background()
	sessions, _ := s.redis.GetAllSessions(ctx, userID)
	for did, session := range sessions {
		if did != currentDeviceID {
			s.redis.AddToBlacklist(ctx, session.RefreshJTI, s.jwtManager.RefreshTTL())
		}
	}

	// 签发新Token
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Name
	}
	accessToken, expiresIn, _ := s.jwtManager.GenerateAccessToken(userID, user.Username, roles, currentDeviceID)
	newRefreshToken, _, newJTI, _ := s.jwtManager.GenerateRefreshToken(userID, currentDeviceID)
	s.redis.SaveSession(ctx, userID, currentDeviceID, &cache.SessionData{RefreshJTI: newJTI, CreatedAt: time.Now().Unix()})

	return &dto.TokenResponse{AccessToken: accessToken, RefreshToken: newRefreshToken, ExpiresIn: expiresIn, TokenType: "Bearer"}, nil
}

func (s *AuthService) generateTokensAndSession(user *model.User, deviceID, ip, userAgent string) (*dto.LoginResponse, error) {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Name
	}
	accessToken, expiresIn, err := s.jwtManager.GenerateAccessToken(user.ID, user.Username, roles, deviceID)
	if err != nil {
		return nil, err
	}
	refreshToken, _, jti, err := s.jwtManager.GenerateRefreshToken(user.ID, deviceID)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	s.redis.SaveSession(ctx, user.ID, deviceID, &cache.SessionData{
		RefreshJTI: jti, UserAgent: userAgent, IP: ip,
		CreatedAt: time.Now().Unix(), LastAccessAt: time.Now().Unix(),
	})
	return &dto.LoginResponse{
		AccessToken: accessToken, RefreshToken: refreshToken,
		ExpiresIn: expiresIn, TokenType: "Bearer", User: dto.ToUserInfo(user),
	}, nil
}

func formatUint64(n uint64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

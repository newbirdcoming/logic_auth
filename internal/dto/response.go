package dto

import "login/internal/model"

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	TokenType    string      `json:"token_type"`
	User         UserInfo    `json:"user"`
}

type UserInfo struct {
	ID       uint64   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Nickname string   `json:"nickname,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type DeviceInfo struct {
	DeviceID    string `json:"device_id"`
	UserAgent   string `json:"user_agent"`
	IP          string `json:"ip"`
	CreatedAt   int64  `json:"created_at"`
	LastAccessAt int64 `json:"last_access_at"`
}

func ToUserInfo(u *model.User) UserInfo {
	roles := make([]string, 0)
	for _, r := range u.Roles {
		roles = append(roles, r.Name)
	}
	return UserInfo{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Nickname:  u.Nickname,
		AvatarURL: u.AvatarURL,
		Roles:     roles,
	}
}

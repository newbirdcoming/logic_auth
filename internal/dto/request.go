package dto

// -------- 请求 ----------

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"password" binding:"required,min=8,max=32"`
	Nickname string `json:"nickname,omitempty" binding:"max=64"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=32"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

type UpdateUserRequest struct {
	Nickname string `json:"nickname,omitempty" binding:"max=64"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	Phone    string `json:"phone,omitempty"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=64"`
	Description string `json:"description,omitempty" binding:"max=256"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name,omitempty" binding:"min=2,max=64"`
	Description string `json:"description,omitempty" binding:"max=256"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []uint64 `json:"permission_ids" binding:"required"`
}

type CreatePermissionRequest struct {
	Code   string `json:"code" binding:"required,min=2,max=128"`
	Name   string `json:"name" binding:"required,min=2,max=128"`
	Module string `json:"module" binding:"required,min=2,max=64"`
}

type UpdateUserStatusRequest struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}

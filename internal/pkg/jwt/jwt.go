package jwt

import (
	"crypto/rsa"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   uint64   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	DeviceID string   `json:"device_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(privateKeyPath, publicKeyPath string, accessTTL, refreshTTL time.Duration) (*JWTManager, error) {
	privData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privData)
	if err != nil {
		return nil, err
	}

	pubData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubData)
	if err != nil {
		return nil, err
	}

	return &JWTManager{
		privateKey: privKey,
		publicKey:  pubKey,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

func (m *JWTManager) GenerateAccessToken(userID uint64, username string, roles []string, deviceID string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(m.accessTTL)
	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    "login-service",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(m.privateKey)
	if err != nil {
		return "", 0, err
	}
	return signed, exp.Unix(), nil
}

func (m *JWTManager) GenerateRefreshToken(userID uint64, deviceID string) (string, int64, string, error) {
	now := time.Now()
	exp := now.Add(m.refreshTTL)
	jti := uuid.New().String()
	claims := Claims{
		UserID:   userID,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    "login-service",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(m.privateKey)
	if err != nil {
		return "", 0, "", err
	}
	return signed, exp.Unix(), jti, nil
}

func (m *JWTManager) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

func (m *JWTManager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *JWTManager) RefreshTTL() time.Duration { return m.refreshTTL }

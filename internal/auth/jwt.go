package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mangooer/gamehub-arena/internal/config"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type JWTService struct {
	cfg *config.JwtConfig
}

func NewJWTService(cfg *config.JwtConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

type Claims struct {
	UserID   uint64   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	Type     string   `json:"type"`
	jwt.RegisteredClaims
}

// 生成访问令牌
func (j *JWTService) GenerateAccessToken(userID uint64, username string, email string, roles []string, tokenType string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.cfg.Issuer,
			Audience:  jwt.ClaimStrings{j.cfg.Audience},
			Subject:   strconv.FormatUint(userID, 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(j.cfg.AccessTokenTTL) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.cfg.SecretKey))
}

// 生产刷新令牌
func (j *JWTService) GenerateRefreshToken(userID uint64, username string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.cfg.Issuer,
			Audience:  jwt.ClaimStrings{j.cfg.Audience},
			Subject:   strconv.FormatUint(userID, 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(j.cfg.RefreshTokenTTL) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.cfg.SecretKey))
}

// 验证令牌
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.cfg.SecretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

// 刷新token
func (j *JWTService) RefreshToken(refreshToken string) (string, error) {
	clams, err := j.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}
	if clams.Type != "refresh" {
		return "", ErrInvalidToken
	}
	accessToken, err := j.GenerateAccessToken(clams.UserID, clams.Username, clams.Email, clams.Roles, "access")
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

package auth

import (
	"errors"
	"unicode"

	"github.com/mangooer/gamehub-arena/internal/config"
	"golang.org/x/crypto/bcrypt"
)

type PasswordService struct {
	cfg *config.PasswordConfig
}

func NewPasswordService(cfg *config.PasswordConfig) *PasswordService {
	return &PasswordService{cfg: cfg}
}

// 验证密码强度
func (p *PasswordService) ValidatePassword(password string) error {
	if len(password) < p.cfg.MinLength {
		return errors.New("密码长度不足")
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsSymbol(char) || unicode.IsPunct(char):
			hasSymbol = true
		}
	}

	if p.cfg.RequireUpper && !hasUpper {
		return errors.New("密码必须包含至少一个大写字母")
	}
	if p.cfg.RequireLower && !hasLower {
		return errors.New("密码必须包含至少一个小写字母")
	}
	if p.cfg.RequireDigit && !hasDigit {
		return errors.New("密码必须包含至少一个数字")
	}
	if p.cfg.RequireSymbol && !hasSymbol {
		return errors.New("密码必须包含至少一个特殊字符")
	}

	return nil
}

// 加密密码
func (p *PasswordService) HashPassword(password string) (string, error) {
	if err := p.ValidatePassword(password); err != nil {
		return "", err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), p.cfg.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil

}

// 验证密码
func (p *PasswordService) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

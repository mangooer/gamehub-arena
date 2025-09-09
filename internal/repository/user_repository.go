package repository

import (
	"github.com/mangooer/gamehub-arena/internal/database"
	"github.com/mangooer/gamehub-arena/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *database.Database
}

func NewUserRepository(db *database.Database) *UserRepository {
	return &UserRepository{db: db}
}

// 创建用户
func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.GetDB().Create(user).Error
}

// 根据ID获取用户
func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	var user models.User
	if err := r.db.GetDB().Preload("UserStats").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.GetDB().Where("username = ?", username).Preload("UserStats").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据邮箱获取用户
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.GetDB().Where("email = ?", email).Preload("UserStats").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 更新用户
func (r *UserRepository) Update(user *models.User) error {
	return r.db.GetDB().Save(user).Error
}

// 删除用户
func (r *UserRepository) Delete(id int64) error {
	return r.db.GetDB().Delete(&models.User{}, id).Error
}

// 获取排行榜
func (r *UserRepository) GetLeaderboard(limit int) ([]models.User, error) {
	var users []models.User
	if err := r.db.GetDB().Order("win_rate DESC,win_count DESC").Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// 更新用户统计
func (r *UserRepository) UpdateStats(userID int64, isWin bool, gameDuration int64) error {
	return r.db.GetDB().Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		if isWin {
			user.WinCount++
		} else {
			user.LoseCount++
		}

		totalGames := user.WinCount + user.LoseCount
		if totalGames > 0 {
			user.WinRate = float64(user.WinCount) / float64(totalGames)
		}

		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		var stats models.UserStats
		if err := tx.Where("user_id = ?", userID).First(&stats).Error; err != nil {
			return err
		}
		stats.TotalGames++
		stats.TotalPlaytime += gameDuration
		if isWin {
			stats.CurrentStreak++
			if stats.CurrentStreak > stats.MaxStreak {
				stats.MaxStreak = stats.CurrentStreak
			}
		} else {
			stats.CurrentStreak = 0
		}

		return tx.Save(&stats).Error
	})
}

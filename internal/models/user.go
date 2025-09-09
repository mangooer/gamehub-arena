package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint64         `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email        string         `json:"email" gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string         `json:"-" gorm:"size:255;not null"`
	Level        int            `json:"level" gorm:"default:1"`
	Experience   int64          `json:"experience" gorm:"default:0"`
	WinCount     int            `json:"win_count" gorm:"default:0"`
	LoseCount    int            `json:"lose_count" gorm:"default:0"`
	WinRate      float64        `json:"win_rate" gorm:"type:decimal(5,4);default:0.0000"`
	Rank         string         `json:"rank" gorm:"size:20;default:'Bronze'"`
	AvatarURL    string         `json:"avatar_url" gorm:"size:255"`
	Status       string         `json:"status" gorm:"size:20;default:'offline'"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系 // 关联关系
	UserStats   *UserStats   `json:"user_stats,omitempty"`
	GameRecords []GameRecord `json:"game_records,omitempty" gorm:"foreignKey:UserID"`
}

type UserStats struct {
	ID            uint64    `json:"id" gorm:"primaryKey"`
	UserID        uint64    `json:"user_id" gorm:"not null;index"`
	TotalGames    int       `json:"total_games" gorm:"default:0"`
	TotalPlaytime int64     `json:"total_playtime" gorm:"default:0"` // 秒
	BestRank      string    `json:"best_rank" gorm:"size:20"`
	CurrentStreak int       `json:"current_streak" gorm:"default:0"`
	MaxStreak     int       `json:"max_streak" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联关系
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// 表名
func (User) TableName() string {
	return "users"
}

func (UserStats) TableName() string {
	return "user_stats"
}

// 钩子函数
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return
}

func (u *User) AfterCreate(tx *gorm.DB) (err error) {
	userStats := UserStats{
		UserID: u.ID,
	}
	return tx.Create(&userStats).Error
}

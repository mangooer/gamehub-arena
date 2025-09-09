package models

import (
	"time"

	"gorm.io/gorm"
)

type GameRoom struct {
	ID             uint64         `json:"id" gorm:"primaryKey"`
	RoomCode       string         `json:"room_code" gorm:"uniqueIndex;size:20;not null"`
	Name           string         `json:"name" gorm:"size:100;not null"`
	MaxPlayers     int            `json:"max_players" gorm:"not null"`
	CurrentPlayers int            `json:"current_players" gorm:"default:0"`
	Status         string         `json:"status" gorm:"size:20;default:'waiting'"`
	GameMode       string         `json:"game_mode" gorm:"size:50;not null"`
	CreatedBy      uint64         `json:"created_by" gorm:"not null;index"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Creator User         `json:"creator" gorm:"foreignKey:CreatedBy"`
	Players []RoomPlayer `json:"players,omitempty"`
	Records []GameRecord `json:"records,omitempty"`
}

type RoomPlayer struct {
	ID       uint64    `json:"id" gorm:"primaryKey"`
	RoomID   uint64    `json:"room_id" gorm:"not null;index"`
	UserID   uint64    `json:"user_id" gorm:"not null;index"`
	Team     string    `json:"team" gorm:"size:10"`
	Position int       `json:"position"`
	JoinedAt time.Time `json:"joined_at"`

	// 关联关系
	Room GameRoom `json:"room" gorm:"foreignKey:RoomID"`
	User User     `json:"user" gorm:"foreignKey:UserID"`
}

type GameRecord struct {
	ID         uint64     `json:"id" gorm:"primaryKey"`
	RoomID     uint64     `json:"room_id" gorm:"not null;index"`
	WinnerTeam string     `json:"winner_team" gorm:"size:10"`
	Duration   int        `json:"duration"` // 游戏时长（秒）
	Status     string     `json:"status" gorm:"size:20;default:'completed'"`
	StartedAt  *time.Time `json:"started_at"`
	EndedAt    *time.Time `json:"ended_at"`
	CreatedAt  time.Time  `json:"created_at"`

	// 关联关系
	Room GameRoom `json:"room" gorm:"foreignKey:RoomID"`
}

func (GameRecord) TableName() string {
	return "game_records"
}

func (GameRoom) TableName() string {
	return "game_rooms"
}

func (RoomPlayer) TableName() string {
	return "room_players"
}

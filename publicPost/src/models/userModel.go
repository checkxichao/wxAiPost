package models

import "time"

type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Password  string    `json:"password" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Power     int       `json:"power" gorm:"power"`
	Show      bool      `json:"show" gorm:"show"`
}

type RefreshToken struct {
	ID        int    `gorm:"primaryKey"`
	UserID    int    `gorm:"index"`
	Token     string `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	CreatedAt time.Time
}
type BlacklistedToken struct {
	ID        int    `gorm:"primaryKey"`
	Token     string `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	CreatedAt time.Time
}

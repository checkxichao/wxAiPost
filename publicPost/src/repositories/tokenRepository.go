// src/repositories/token_repository.go

package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"publicPost/src/models"
)

type TokenRepository interface {
	SaveRefreshToken(userID int, token string, expiry time.Time) error
	IsRefreshTokenValid(userID int, token string) (bool, error)
	RevokeRefreshToken(token string) error
	BlacklistToken(token string, expiry time.Time) error
	IsTokenBlacklisted(token string) (bool, error)
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) SaveRefreshToken(userID int, token string, expiry time.Time) error {
	refreshToken := &models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiry,
	}
	return r.db.Create(refreshToken).Error
}

func (r *tokenRepository) IsRefreshTokenValid(userID int, token string) (bool, error) {
	var refreshToken models.RefreshToken
	err := r.db.Where("user_id = ? AND token = ? AND expires_at > ?", userID, token, time.Now()).First(&refreshToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *tokenRepository) RevokeRefreshToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.RefreshToken{}).Error
}

func (r *tokenRepository) BlacklistToken(token string, expiry time.Time) error {
	blacklistedToken := &models.BlacklistedToken{
		Token:     token,
		ExpiresAt: expiry,
	}
	return r.db.Create(blacklistedToken).Error
}

func (r *tokenRepository) IsTokenBlacklisted(token string) (bool, error) {
	var blacklistedToken models.BlacklistedToken
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&blacklistedToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

package repositories

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"publicPost/src/models"
)

type GptRepository interface {
	GetGPTInfo() (*models.GptModel, error)
	SetGPTInfo(gpt models.GptModel) error
}
type gptRepository struct {
	db *gorm.DB
}

func NewGptRepository(db *gorm.DB) GptRepository {
	return &gptRepository{db: db}
}

func (r *gptRepository) GetGPTInfo() (*models.GptModel, error) {
	var gptInfo models.GptModel

	if err := r.db.Table("gpt_setting").Select("*").First(&gptInfo).Error; err != nil {
		return nil, err // 返回错误
	}

	return &gptInfo, nil
}

func (r *gptRepository) SetGPTInfo(gpt models.GptModel) error {

	if gpt.Key == "" || gpt.Model == "" {
		return errors.New("key或model传入不正确")
	}

	updates := map[string]interface{}{
		"key":   gpt.Key,
		"model": gpt.Model,
	}

	// 执行更新操作
	err := r.db.Table("gpt_setting").Where("id = 1").Updates(updates).Error
	if err != nil {
		// 如果更新失败，返回错误
		return fmt.Errorf("更新失败: %v", err)
	}

	return nil
}

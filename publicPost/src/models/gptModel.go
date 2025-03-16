package models

type GptModel struct {
	Id    int    `json:"id" gorm:"not null"`
	Key   string `json:"key" gorm:"unique;not null"`
	Model string `json:"model" gorm:"not null"`
}

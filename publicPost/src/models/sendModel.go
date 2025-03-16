package models

import "time"

type Article struct {
	Title        string `json:"title"`
	Author       string `json:"author"`
	Content      string `json:"content"`
	ThumbMediaId string `json:"thumb_media_id"`
}
type DraftRequest struct {
	Articles []Article `json:"articles"`
}

type ThumbMediaModel struct {
	Id           int       `json:"id" gorm:"primaryKey"`
	ThumbMediaId string    `json:"thumbMediaId" gorm:"column:thumb_media_id"`
	Wxid         string    `json:"wxid" gorm:"column:wxid"`
	Note         string    `json:"note" gorm:"column:note"`
	ImgUrl       string    `json:"imgUrl" gorm:"column:img_url"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type TitleListModel struct {
	Id        int       `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"column:title"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
	State     bool      `json:"state" gorm:"column:state"`
}
type ScheduledTask struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	WxID             string    `gorm:"column:wxid;size:255;not null" json:"wxid"`
	Secret           string    `gorm:"size:255;not null" json:"secret"`
	ScheduledTime    time.Time `gorm:"not null" json:"scheduledTime"`
	TaskType         string    `gorm:"size:50;not null" json:"taskType"`
	Mode             string    `gorm:"size:50;not null" json:"mode"`
	ArticleCount     int       `gorm:"not null" json:"articleCount"`
	SelectedArticles string    `gorm:"type:JSON;not null" json:"selectedArticles"`
	SelectedTitles   string    `gorm:"type:JSON" json:"selectedTitles"`
	ThumbID          string    `gorm:"size:255" json:"thumbId"`
	TemplateID       string    `gorm:"size:255" json:"templateId"`
	MediaID          string    `gorm:"size:255" json:"mediaId"`
	Status           string    `gorm:"size:50;default:'pending'" json:"status"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}

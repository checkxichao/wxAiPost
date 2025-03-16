package models

import "time"

type WechatAccount struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Wxid        string    `json:"wxid" gorm:"unique;not null"`
	Secret      string    `json:"secret" gorm:"not null"`
	IsUse       bool      `json:"isuse" gorm:"not null"`
	BindWechat  string    `json:"bindWechat" gorm:"not null"`
	Show        bool      `json:"show" gorm:"not null"`
	NowState    int       `json:"nowState" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	AccessToken string    `json:"access_token"`
	GetTime     time.Time `json:"get_time" `
	EndTime     time.Time `json:"end_time"`
}

type WechatMediaModel struct {
	Id         int       `json:"id" gorm:"column:id;primary_key"`
	MediaId    string    `json:"mediaId" gorm:"column:media_id"`
	BelongWxid string    `json:"belongWxid" gorm:"column:belong_wxid"`
	State      bool      `json:"state" gorm:"column:state"`
	Title      string    `json:"title" gorm:"column:title"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
type WechatData struct {
	Id         int    `json:"id" gorm:"column:id;primary_key"`
	WechatName string `json:"wechatName" gorm:"column:wechat_name"`
	Follow     int    `json:"follow" gorm:"column:follow"`
	Unfollow   int    `json:"unfollow" gorm:"column:unfollow"`
	Punish     int    `json:"punish" gorm:"column:punish"`
}

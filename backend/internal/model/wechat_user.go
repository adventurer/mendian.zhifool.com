package model

import "time"

type WeChatUser struct {
	ID        uint      `json:"-" gorm:"primaryKey;autoIncrement"`
	AppID     string    `json:"-" gorm:"column:app_id;size:64;not null;uniqueIndex:idx_wechat_users_app_openid,priority:1"`
	OpenID    string    `json:"-" gorm:"column:openid;size:128;not null;uniqueIndex:idx_wechat_users_app_openid,priority:2"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (WeChatUser) TableName() string {
	return "wechat_users"
}

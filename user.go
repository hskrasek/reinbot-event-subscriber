package main

type User struct {
	Id string `json:"id" gorm:"column:slack_user_id"`
	Name string `json:"name" gorm:"column:username"`
	TimeZone string `gorm:"column:timezone"`
}

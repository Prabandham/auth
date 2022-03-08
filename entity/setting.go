package entity

type Setting struct {
	Base
	DarkTheme bool   `json:"DarkTheme"`
	UserId    string `gorm:"index;not null" binding:"required" json:"UserId"`
	User      User
}

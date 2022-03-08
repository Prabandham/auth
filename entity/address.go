package entity

type Address struct {
	Base
	Line1                      string `gorm:"size:1000" json:"Line1" sql:"not null"`
	Line2                      string `gorm:"size:1000" json:Line2" sql:"not null"`
	State                      string `gorm:"size:256" json:State" sql:"not null"`
	City                       string `gorm:"size:256" json:City" sql:"not null"`
	PinCode                    string `gorm:"size:256" json:PinCode" sql:"not null"`
	CurrentResendentialAddress bool   `json:"CurrentResendentialAddress"`
}

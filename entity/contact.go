package entity

type Contact struct {
	Base
	PrimaryMobile    string `gorm:index;size:256" json:"PrimaryMobile"`
	AlternateMobile  string `gorm:index;size:256" json:"AlternateMobile"`
	EmergencyContact string `gorm:index:size:256" json:"EmergencyContact" sql:"not null"`
}

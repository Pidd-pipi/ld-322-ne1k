package model

import "time"

type Schedule struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	DeviceID  uint       `gorm:"index" json:"deviceId"`
	Hour      int        `json:"hour"`
	Minute    int        `json:"minute"`
	Action    string     `gorm:"type:varchar(32)" json:"action"`
	Enabled   bool       `json:"enabled"`
	LastRunAt *time.Time `json:"lastRunAt"`
	CreatedAt time.Time  `json:"createdAt"`
	Device    *Device    `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
}

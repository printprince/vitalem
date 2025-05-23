package models

import "time"

const (
	RoleAdmin   = "admin"
	RoleDoctor  = "doctor"
	RolePatient = "patient"
)

type Users struct {
	ID             uint   `gorm:"primary_key;auto_increment" json:"id"`
	Email          string `gorm:"type:varchar(255);unique_index"`
	HashedPassword string `gorm:"size:255"`
	Role           string
	CreatedAt      time.Time
}

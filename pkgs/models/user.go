package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `json:"email" gorm:"not null;uniqueIndex"`
	Password string `json:"-" gorm:"not null"`
	RoleID   uint   `json:"role_id" gorm:"not null"`
	Role     Role   `json:"role" gorm:"foreignKey:RoleID"`
}

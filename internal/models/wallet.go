package models

import (
	"time"
)

type Wallet struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"-"`
	Balance   int64     `gorm:"not null;default:0" json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

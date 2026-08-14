package models

import (
	"time"
)

type Transaction struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	WalletID        uint      `gorm:"index;not null" json:"wallet_id"`
	Type            string    `gorm:"type:varchar(20);not null" json:"type"`
	Amount          int64     `gorm:"not null" json:"amount"`
	Category        string    `gorm:"type:varchar(50)" json:"category"`
	Note            string    `json:"note"`
	RelatedWalletID *uint     `json:"related_wallet_id"`
	CreatedAt       time.Time `json:"created_at"`
}

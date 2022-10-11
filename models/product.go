package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ID        uuid.UUID `json:"ID"`
	Available int       `json:"available" binding:"min=1,max=99"`
	Price     int       `json:"price" binding:"min=0,max=1000"`
	Name      string    `json:"name" gorm:"unique" binding:"min=2,max=30"`
	SellerID  uuid.UUID `json:"seller_id"`
}

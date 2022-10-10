package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	UUID   uuid.UUID
	Valid  bool
	UserID uuid.UUID
}

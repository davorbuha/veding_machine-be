package models

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	Buyer = iota
	Seller
)

type User struct {
	gorm.Model
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name" binding:"required,alpha,min=5,max=20"`
	Username string    `json:"username" gorm:"unique" binding:"required,alphanum,min=5,max=20"`
	Password string    `json:"password" binding:"required,alphanum,min=5,max=20"`
	Role     int       `json:"role" binding:"eq=0|eq=1"`
}

func (user *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	user.Password = string(bytes)
	return nil
}

func (user *User) CheckPassword(providedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(providedPassword))
	if err != nil {
		return err
	}
	return nil
}

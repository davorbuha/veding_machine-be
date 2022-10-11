package models

import "github.com/google/uuid"

type Balance struct {
	UserID  uuid.UUID `json:"user_id"`
	FIVE    int
	TEN     int
	TWENTY  int
	FIFTY   int
	HUNDRED int
}

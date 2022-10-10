package auth

import (
	"errors"
	"mvpmatch/veding-machine/database"
	"mvpmatch/veding-machine/models"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var jwtKey = []byte("supersecretkey")

type JWTClaim struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Session  uuid.UUID `json:"session"`
	Role     int       `json:"role"`
	jwt.StandardClaims
}

func GenerateAccessJWT(username string, userId uuid.UUID, role int, session uuid.UUID) (tokenString string, err error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &JWTClaim{
		UserID:   userId,
		Username: username,
		Session:  session,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(jwtKey)
	return
}

func GenerateRefreshJWT(username string, userId uuid.UUID, sessionUUID uuid.UUID) (tokenString string, err error) {
	expirationTime := time.Now().Add(24 * 365 * time.Hour)
	claims := &JWTClaim{
		UserID:   userId,
		Username: username,
		Session:  sessionUUID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	session := models.Session{
		UserID: userId,
		UUID:   sessionUUID,
		Valid:  true,
	}
	record := database.Instance.Create(&session)
	if record.Error != nil {
		return "", record.Error
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(jwtKey)
	return
}

func GetClaimsFromToken(signedToken string) (claims *JWTClaim, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}

	return
}

func ValidateAccessToken(signedToken string) (err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return
	}

	session := models.Session{}
	record := database.Instance.Where("uuid = ?", claims.Session).First(&session)
	if record.Error != nil {
		return record.Error
	}

	if !session.Valid {
		return errors.New("session invalid")
	}

	return
}

func ValidateRefreshToken(signedToken string) (user models.User, sessionUUID uuid.UUID, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return
	}

	session := models.Session{}
	record := database.Instance.Where("uuid = ?", claims.Session).First(&session)
	if record.Error != nil {
		return user, uuid.UUID{}, record.Error
	}

	if !session.Valid {
		return user, uuid.UUID{}, errors.New("session invalid")
	}

	user.Username = claims.Username
	user.ID = claims.UserID
	sessionUUID = claims.Session

	return
}

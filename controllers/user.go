package controllers

import (
	"errors"
	"mvpmatch/veding-machine/auth"
	"mvpmatch/veding-machine/database"
	"mvpmatch/veding-machine/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func RegisterUser(context *gin.Context) {
	user := models.User{}
	if err := context.ShouldBindJSON(&user); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make([]ErrorMsg, len(ve))
			for i, fe := range ve {
				out[i] = ErrorMsg{fe.Field(), getErrorMsg(fe)}
			}
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": out})
		}
		return
	}

	if err := user.HashPassword(user.Password); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	user.ID = uuid.New()
	record := database.Instance.Create(&user)
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}

	context.JSON(http.StatusCreated, gin.H{"userId": user.ID, "username": user.Username})
	return
}

type TokenRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=5,max=20"`
	Password string `json:"password" binding:"required,alphanum,min=5,max=20"`
}

func Login(context *gin.Context) {
	var request TokenRequest
	var user models.User
	if err := context.ShouldBindJSON(&request); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make([]ErrorMsg, len(ve))
			for i, fe := range ve {
				out[i] = ErrorMsg{fe.Field(), getErrorMsg(fe)}
			}
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": out})
		}
		return
	}

	// check if username exists and password is correct
	record := database.Instance.Where("username = ?", request.Username).First(&user)
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}

	credentialError := user.CheckPassword(request.Password)
	if credentialError != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		context.Abort()
		return
	}

	sessionUUID := uuid.New()

	// generate tokens
	accessToken, err := auth.GenerateAccessJWT(user.Username, user.ID, user.Role, sessionUUID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	refreshToken, err := auth.GenerateRefreshJWT(user.Username, user.ID, sessionUUID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{"at": accessToken, "rt": refreshToken})
	return
}

type RefreshTokenRequest struct {
	RT string `json:"rt"`
}

func RefreshToken(context *gin.Context) {
	rq := RefreshTokenRequest{}
	if err := context.ShouldBindJSON(&rq); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	user, sessionUUID, err := auth.ValidateRefreshToken(rq.RT)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	// generate tokens
	accessToken, err := auth.GenerateAccessJWT(user.Username, user.ID, user.Role, sessionUUID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{"at": accessToken})
	return
}

func Logout(context *gin.Context) {
	at := context.GetHeader("Authorization")

	claims, err := auth.GetClaimsFromToken(at)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		context.Abort()
		return
	}

	database.Instance.Model(&models.Session{}).Where("uuid = ?", claims.Session).Update("valid", false)
	context.JSON(http.StatusOK, gin.H{"ok": true})
	context.Abort()
	return
}

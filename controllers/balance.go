package controllers

import (
	"errors"
	"mvpmatch/veding-machine/auth"
	"mvpmatch/veding-machine/database"
	"mvpmatch/veding-machine/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type DepositRequest struct {
	Amount int `json:"amount" binding:"eq=5|eq=10|eq=20|eq=50|eq=100"`
}

func Deposit(context *gin.Context) {
	token := auth.GetToken(context)
	var deposit DepositRequest
	if err := context.ShouldBindJSON(&deposit); err != nil {
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

	claims, err := auth.GetClaimsFromToken(token)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	balance := &models.Balance{UserID: claims.UserID}
	record := database.Instance.Where("user_id = ? ", claims.UserID).First(&balance)
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}

	switch deposit.Amount {
	case 5:
		balance.FIVE++
	case 10:
		balance.TEN++
	case 20:
		balance.TWENTY++
	case 50:
		balance.FIFTY++
	case 100:
		balance.HUNDRED++
	}

	record = database.Instance.Where("user_id = ? ", claims.UserID).Save(&balance)
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"ok": true})
}

func ResetDeposit(context *gin.Context) {
	token := auth.GetToken(context)
	claims, err := auth.GetClaimsFromToken(token)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	record := database.Instance.Model(&models.Balance{}).Where("user_id = ? ", claims.UserID).Update("FIVE", 0).Update("TEN", 0).Update("TWENTY", 0).Update("FIFTY", 0).Update("HUNDRED", 0)
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"ok": true})
}

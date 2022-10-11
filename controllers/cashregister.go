package controllers

import (
	"errors"
	"math"
	"mvpmatch/veding-machine/auth"
	"mvpmatch/veding-machine/database"
	"mvpmatch/veding-machine/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	FIVE = iota
	TEN
	TWENTY
	FIFTY
	HUNDRED
)

type BuyRequest struct {
	ProductId uuid.UUID `json:"product_id"`
	Amount    int       `json:"amount" binding:"gte=1,lte=100"`
}

func Buy(context *gin.Context) {
	token := auth.GetToken(context)

	var buy BuyRequest
	if err := context.ShouldBindJSON(&buy); err != nil {
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

	user := models.User{}
	record := database.Instance.Where("id = ? ", claims.UserID).First(&user)
	if record.Error != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		return
	}

	product := models.Product{}
	record = database.Instance.Where("id = ? ", buy.ProductId).First(&product)
	if record.Error != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		return
	}

	balance := models.Balance{}
	record = database.Instance.Where("user_id = ? ", claims.UserID).First(&balance)
	if record.Error != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		return
	}

	summedBalance := sumBalance(balance)

	if product.Price*buy.Amount > summedBalance {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "not enough money"})
		context.Abort()
		return
	}

	if buy.Amount > product.Available {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "not enough products"})
		context.Abort()
		return
	}

	cost := product.Price * buy.Amount

	balanceToReturn := calculateChange(cost, balance)

	product.Available -= buy.Amount

	record = database.Instance.Where("id = ? ", buy.ProductId).Save(&product)
	if record.Error != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		return
	}
	record = database.Instance.Where("user_id = ? ", claims.UserID).Save(&balanceToReturn)
	if record.Error != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		return
	}

	balanceArray := []int{}

	for i := 0; i < balanceToReturn.HUNDRED; i++ {
		balanceArray = append(balanceArray, 100)
	}
	for i := 0; i < balanceToReturn.FIFTY; i++ {
		balanceArray = append(balanceArray, 50)
	}
	for i := 0; i < balanceToReturn.TWENTY; i++ {
		balanceArray = append(balanceArray, 20)
	}
	for i := 0; i < balanceToReturn.TEN; i++ {
		balanceArray = append(balanceArray, 10)
	}
	for i := 0; i < balanceToReturn.FIVE; i++ {
		balanceArray = append(balanceArray, 5)
	}

	context.JSON(http.StatusOK, gin.H{"spent": cost, "change": balanceArray})
}

func sumBalance(b models.Balance) int {
	sum := b.FIVE*5 + b.TEN*10 + b.TWENTY*20 + b.FIFTY*50 + b.HUNDRED*100
	return sum
}

func calculateChange(cost int, b models.Balance) (fb models.Balance) {
	sum := sumBalance(b)
	change := sum - cost

	for i := 4; i >= 0; i-- {
		if change == 0 {
			return
		}
		switch i {
		case HUNDRED:
			fb.HUNDRED = int(math.Floor(float64(change) / 100))
			change -= fb.HUNDRED * 100
		case FIFTY:
			fb.FIFTY = int(math.Floor(float64(change) / 50))
			change -= fb.FIFTY * 50
		case TWENTY:
			fb.TWENTY = int(math.Floor(float64(change) / 20))
			change -= fb.TWENTY * 20
		case TEN:
			fb.TEN = int(math.Floor(float64(change) / 10))
			change -= fb.TEN * 10
		case FIVE:
			fb.FIVE = int(math.Floor(float64(change) / 5))
			change -= fb.FIVE * 5
		}
	}
	return
}

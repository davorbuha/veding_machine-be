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

func CreateProduct(context *gin.Context) {
	token := auth.GetToken(context)
	product := models.Product{}
	if err := context.ShouldBindJSON(&product); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make([]ErrorMsg, len(ve))
			for i, fe := range ve {
				out[i] = ErrorMsg{fe.Field(), getErrorMsg(fe)}
			}
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": out})
		} else {
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	if product.Price%5 != 0 {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "price not divisible by 5 or 10"})
		return
	}

	product.ID = uuid.New()

	claims, err := auth.GetClaimsFromToken(token)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	product.SellerID = claims.UserID

	record := database.Instance.Create(&product)

	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"product_id": product.ID})
}

func UpdateProduct(context *gin.Context) {
	token := auth.GetToken(context)
	product := models.Product{}
	if err := context.ShouldBindJSON(&product); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make([]ErrorMsg, len(ve))
			for i, fe := range ve {
				out[i] = ErrorMsg{fe.Field(), getErrorMsg(fe)}
			}
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": out})
		} else {
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	if product.ID == uuid.Nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing product id"})
		context.Abort()
		return
	}

	claims, err := auth.GetClaimsFromToken(token)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	product.SellerID = claims.UserID

	record := database.Instance.Where("id = ? AND seller_id = ?", product.ID, product.SellerID).Update("available", product.Available).Update("price", product.Price).Update("name", product.Name)
	if record.RowsAffected == 0 {
		context.JSON(http.StatusForbidden, gin.H{"error": "you dont have permissions to update this product"})
		return
	}
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"product_id": product.ID})
}

type DeleteProductRequest struct {
	ID uuid.UUID `json:"id"`
}

func DeleteProduct(context *gin.Context) {
	token := auth.GetToken(context)
	rq := DeleteProductRequest{}
	err := context.ShouldBindJSON(&rq)

	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := auth.GetClaimsFromToken(token)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	record := database.Instance.Where("id = ? AND seller_id = ?", rq.ID, claims.UserID).Delete(&models.Product{})
	if record.RowsAffected == 0 {
		context.JSON(http.StatusForbidden, gin.H{"error": "you dont have permissions to delete this product"})
		return
	}
	if record.Error != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"product_id": rq.ID})
}

func GetProducts(context *gin.Context) {
	products := []models.Product{}
	record := database.Instance.Find(&products)
	if record.Error != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"products": products})
}

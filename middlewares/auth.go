package middlewares

import (
	"mvpmatch/veding-machine/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
		if tokenString == "" {
			context.JSON(401, gin.H{"error": "request does not contain an access token"})
			context.Abort()
			return
		}
		err := auth.ValidateAccessToken(tokenString)
		if err != nil {
			context.JSON(401, gin.H{"error": err.Error()})
			context.Abort()
			return
		}
		context.Next()
	}
}

func RoleGuard(role int) gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
		err := auth.ValidateAccessToken(tokenString)
		if err != nil {
			context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			context.Abort()
			return
		}
		claims, err := auth.GetClaimsFromToken(tokenString)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			context.Abort()
			return
		}

		if claims.Role != role {
			context.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			context.Abort()
			return
		}

		context.Next()
	}
}

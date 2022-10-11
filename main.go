package main

import (
	"mvpmatch/veding-machine/config"
	"mvpmatch/veding-machine/controllers"
	"mvpmatch/veding-machine/database"
	"mvpmatch/veding-machine/middlewares"
	"mvpmatch/veding-machine/models"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	c := config.Config{}
	err := env.Parse(&c)
	if err != nil {
		panic(err)
	}

	database.Connect(c.DSN)
	database.Migrate()

	// Initialize Router
	router := initRouter()
	router.Run(c.Port)

}

func initRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/ping", controllers.Ping)
		api.POST("/login", controllers.Login)
		api.POST("/refresh-token", controllers.RefreshToken)
		api.POST("/user/register", controllers.RegisterUser)
		api.POST("/logout-all", controllers.LogoutAll)
		secured := api.Group("/secured").Use(middlewares.Auth())
		{
			secured.GET("/ping", controllers.Ping)
			secured.POST("/logout", controllers.Logout)
			secured.PUT("/product", middlewares.RoleGuard(models.Seller), controllers.CreateProduct)
			secured.DELETE("/product", middlewares.RoleGuard(models.Seller), controllers.DeleteProduct)
			secured.POST("/product", middlewares.RoleGuard(models.Seller), controllers.UpdateProduct)
			secured.POST("/deposit", middlewares.RoleGuard(models.Buyer), controllers.Deposit)
			secured.POST("/reset-deposit", middlewares.RoleGuard(models.Buyer), controllers.ResetDeposit)
			secured.POST("/buy", middlewares.RoleGuard(models.Buyer), controllers.Buy)
		}
		api.GET("/products", controllers.GetProducts)
	}
	return router
}

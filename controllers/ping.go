package controllers

import "github.com/gin-gonic/gin"

func Ping(ctx *gin.Context) {
	ctx.Writer.Write([]byte("pong"))
}

package main

import (
    "github.com/gin-gonic/gin"
    "log"
    "os"
)

func main() {
    router := gin.Default()
    router.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
    log.Println("Server running on port 8080")
    router.Run(":8080")
}

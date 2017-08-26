package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func main() {
    router := gin.Default()
    router.LoadHTMLGlob("frontend/*.html")
    router.Static("/dist", "./frontend/dist")
    router.GET("/", rootHandler)
    router.Run()
}

func rootHandler(c *gin.Context) {
    // c.JSON(200, gin.H{
    //     "message": "pong",
    // })
    c.HTML(http.StatusOK, "index.html", nil)
}

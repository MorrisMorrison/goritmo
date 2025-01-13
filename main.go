package main

import (
	"log"
	"net/http"

	"github.com/MorrisMorrison/goritmo/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main(){
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"*"}, 
        AllowMethods: []string{
            http.MethodGet,
            http.MethodHead,
            http.MethodPut,
            http.MethodPatch,
            http.MethodPost,
            http.MethodDelete,
        },
        AllowHeaders: []string{
            echo.HeaderOrigin,
            echo.HeaderContentType,
            echo.HeaderAccept,
            echo.HeaderAuthorization,
        },
        AllowCredentials: true,
    }))

    e.GET("/ws", api.HandleWebSocket)
    e.GET("/rooms", api.ListRooms)
	e.POST("/rooms", api.CreateRoom)
    e.GET("/health", api.HealthCheck)

    // Start server
    log.Println("Starting signaling server on :8080")
    e.Logger.Fatal(e.Start(":8080"))
}
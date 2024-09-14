package main

import (
	"log"
	"net/http"

	"github.com/FiveEightyEight/mwfapi/db"
	"github.com/FiveEightyEight/mwfapi/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

func homePath(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome to Math With Friends")
}

func main() {

	// PORT: 6379
	// PID: 66060
	opts := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	rdb := db.NewRedisClient(opts)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://192.168.1.155:5173"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	e.GET("/", homePath)
	e.POST("/login", handlers.Login(rdb))
	e.POST("/refresh", handlers.RefreshToken)

	gameGroup := e.Group("/v1/api")
	gameGroup.Use(handlers.AuthMiddleware)
	gameGroup.GET("/game/:game_session_id", handlers.ConnectToGameSession(rdb))
	// gameGroup.GET("/:id", handlers.GetGame(rdb))
	port := ":8088"
	e.Logger.Fatal(e.Start(port))
	log.Printf("Server is running on port %s\n", port)
}

package main

import (
	"log"
	"net/http"

	"github.com/FiveEightyEight/mwfapi/db"
	"github.com/FiveEightyEight/mwfapi/handlers"
	"github.com/gorilla/websocket"
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

	// WebSocket configuration
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins, in production ensureproper origin checking.
		},
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://192.168.1.155:5173"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodConnect, http.MethodHead, http.MethodPatch, http.MethodTrace, http.MethodTrace},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderUpgrade, echo.HeaderConnection, "credentials"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Set-Cookie", "Access-Control-Allow-Origin"},
	}))

	e.OPTIONS("/*", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	e.GET("/", homePath)
	e.POST("/login", handlers.Login(rdb))
	e.POST("/register", handlers.Register(rdb))
	e.POST("/refresh", handlers.RefreshToken)
	e.GET("/active-sessions", handlers.GetActiveGameSessions(rdb))

	gameGroup := e.Group("/v1/api")
	gameGroup.Use(handlers.AuthMiddleware)
	gameGroup.POST("/game/create", handlers.CreateGame(rdb))
	gameGroup.GET("/game/:game_session_id", handlers.ConnectToGameSession(rdb, upgrader))

	port := ":8088"
	e.Logger.Fatal(e.Start("0.0.0.0" + port))
	log.Printf("Server is running on port %s\n", port)
}

package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func homePath(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome to Math With Friends")
}

func main() {
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

	port := ":8088"
	e.Logger.Fatal(e.Start(port))
	log.Printf("Server is running on port %s\n", port)
}

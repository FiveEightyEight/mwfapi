package handlers

import (
	"net/http"
	"time"

	"strings"

	"github.com/FiveEightyEight/mwfapi/auth"
	"github.com/FiveEightyEight/mwfapi/db"
	"github.com/FiveEightyEight/mwfapi/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	accessTokenCookieName  = "t"
	refreshTokenCookieName = "mt"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check for the refresh token cookie
		cookie, err := c.Cookie(refreshTokenCookieName)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Refresh token cookie is missing"})
		}
		refreshToken := cookie.Value

		// Validate the refresh token
		claims, err := auth.ValidateToken(refreshToken, true)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid refresh token"})
		}

		// Set the user ID in the context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		return next(c)
	}
}

func RefreshToken(c echo.Context) error {
	cookie, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Refresh token cookie is missing"})
	}
	refreshToken := cookie.Value

	newAccessToken, newRefreshToken, err := auth.RefreshTokens(refreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid refresh token"})
	}

	c.SetCookie(createRefreshTokenCookie(newRefreshToken))

	return c.JSON(http.StatusOK, map[string]string{
		"t": newAccessToken,
	})
}

func Login(rdb *db.RedisClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(refreshTokenCookieName)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Refresh token cookie is missing"})
		}
		refreshToken := cookie.Value

		// Validate the refresh token
		claims, err := auth.ValidateToken(refreshToken, true)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid refresh token"})
		}

		// Retrieve user from Redis cache

		user, err := rdb.GetUser(c.Request().Context(), uuid.MustParse(claims.UserID))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user"})
		}

		// Generate new refresh token
		newRefreshToken, err := auth.GenerateRefreshToken(user.ID.String(), user.Username)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate new refresh token"})
		}

		// Set the new refresh token as a cookie
		c.SetCookie(createRefreshTokenCookie(newRefreshToken))

		// Generate new access token
		newAccessToken, err := auth.GenerateAccessToken(user.ID.String(), user.Username)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate new access token"})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"t": newAccessToken,
		})
	}
}

func Register(rdb *db.RedisClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Username string `json:"username"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		}

		// Validate username length
		if len(req.Username) < 3 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username must be at least 3 characters long"})
		}

		// Create a new user
		newUser := &models.User{
			ID:       uuid.New(),
			Username: strings.ToLower(req.Username),
		}

		// Save the user to Redis
		err := rdb.CreateUser(c.Request().Context(), newUser)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
		}

		// Generate refresh token
		refreshToken, err := auth.GenerateRefreshToken(newUser.ID.String(), newUser.Username)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate refresh token"})
		}

		// Set the refresh token as a cookie
		c.SetCookie(createRefreshTokenCookie(refreshToken))

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"user_id":  newUser.ID,
			"username": newUser.Username,
		})
	}
}

func createRefreshTokenCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		Domain:   "192.168.1.155",
		HttpOnly: true,
		Secure:   false, // Set to true in production (over HTTPS)
		SameSite: http.SameSiteLaxMode,
	}
}

package handlers

import (
	"net/http"
	"time"

	"github.com/FiveEightyEight/mwfapi/auth"
	"github.com/FiveEightyEight/mwfapi/db"
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

	c.SetCookie(&http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    newRefreshToken,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,                 // Ensure this is true in production (over HTTPS)
		SameSite: http.SameSiteNoneMode, // Ensure this is true in production (over HTTPS)
	})

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
		c.SetCookie(&http.Cookie{
			Name:     refreshTokenCookieName,
			Value:    newRefreshToken,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
			HttpOnly: true,
			Secure:   true, // Ensure this is true in production (over HTTPS)
			SameSite: http.SameSiteStrictMode,
		})

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

package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/FiveEightyEight/mwfapi/auth"
	"github.com/FiveEightyEight/mwfapi/db"
	"github.com/FiveEightyEight/mwfapi/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CreateUserRequest struct {
	Username string `json:"username"`
}

func CreateUser(rdb *db.RedisClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req CreateUserRequest
		if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		}

		if req.Username == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username is required"})
		}

		newUser := &models.User{
			ID:       uuid.New(),
			Username: strings.ToLower(req.Username),
		}

		err := rdb.CreateUser(c.Request().Context(), newUser)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
		}

		refreshToken, err := auth.GenerateRefreshToken(newUser.ID.String(), newUser.Username)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate refresh token"})
		}

		c.SetCookie(createRefreshTokenCookie(refreshToken))

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"user_id":       newUser.ID,
			"username":      newUser.Username,
			"refresh_token": refreshToken,
		})
	}
}

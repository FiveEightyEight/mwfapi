package handlers

import (
	"net/http"

	"github.com/FiveEightyEight/mwfapi/db"
	"github.com/FiveEightyEight/mwfapi/game"
	"github.com/FiveEightyEight/mwfapi/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now. Change in production :)
	},
}

func CreateGame(rdb *db.RedisClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the game name from the request
		var req struct {
			Name       string            `json:"name"`
			GameConfig models.GameConfig `json:"game_config"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}

		// Check if the game name is provided
		if req.Name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Game name is required"})
		}

		// Create a new game
		newGame := &models.Game{
			ID:   uuid.New(),
			Name: req.Name,
		}

		// Save the game to Redis
		err := rdb.CreateGame(c.Request().Context(), newGame)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create game"})
		}

		// Generate game problems
		problems := game.GenerateGameProblems(req.GameConfig)

		// Create a new game session
		gameSession := &models.GameSession{
			ID:                  uuid.New(),
			Name:                req.Name,
			GameID:              newGame.ID,
			Status:              "waiting",
			Players:             []models.User{},
			Scores:              []models.Score{},
			GameConfig:          req.GameConfig,
			Problems:            problems,
			CurrentProblemIndex: 0,
		}

		// Save the game session to Redis
		err = rdb.CreateGameSession(c.Request().Context(), gameSession)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create game session"})
		}

		// Upgrade the HTTP connection to a WebSocket connection
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to upgrade to WebSocket"})
		}
		defer ws.Close()

		// Subscribe to game session updates
		updates, err := rdb.SubscribeToGameSession(c.Request().Context(), gameSession.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to subscribe to game session"})
		}

		// Send initial game session data
		err = ws.WriteJSON(gameSession)
		if err != nil {
			return nil
		}

		// Listen for updates and send them to the client
		for update := range updates {
			err = ws.WriteJSON(update)
			if err != nil {
				break
			}
		}

		return nil
	}
}

func GetActiveGameSessions(rdb *db.RedisClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		activeSessions, err := rdb.GetActiveGameSessions(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve active game sessions"})
		}
		return c.JSON(http.StatusOK, activeSessions)
	}
}

func CloseGameSession(rdb *db.RedisClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionID := c.Param("id")
		if sessionID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Session ID is required"})
		}

		err := rdb.CloseGameSession(c.Request().Context(), uuid.MustParse(sessionID))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to close game session"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Game session closed successfully"})
	}
}

func UpdateGameSession(rdb *db.RedisClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionID := c.Param("id")
		if sessionID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Session ID is required"})
		}

		var updatedSession models.GameSession
		if err := c.Bind(&updatedSession); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		}

		updatedSession.ID = uuid.MustParse(sessionID)
		err := rdb.UpdateGameSession(c.Request().Context(), &updatedSession)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update game session"})
		}

		return c.JSON(http.StatusOK, updatedSession)
	}
}

func ConnectToGameSession(rdb *db.RedisClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionID := c.Param("id")
		if sessionID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Session ID is required"})
		}

		// Upgrade the HTTP connection to a WebSocket connection
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to upgrade to WebSocket"})
		}
		defer ws.Close()

		// Subscribe to game session updates
		updates, err := rdb.SubscribeToGameSession(c.Request().Context(), uuid.MustParse(sessionID))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to subscribe to game session"})
		}

		// Send initial game session data
		initialSession, err := rdb.GetGameSession(c.Request().Context(), uuid.MustParse(sessionID))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get initial game session data"})
		}
		err = ws.WriteJSON(initialSession)
		if err != nil {
			return nil
		}

		// Listen for updates and send them to the client
		for update := range updates {
			err = ws.WriteJSON(update)
			if err != nil {
				break
			}
		}

		return nil
	}
}

package handlers

import (
	"net/http"

	"context"
	"log"

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

		return c.JSON(http.StatusCreated, gameSession)
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
		sessionID := c.Param("game_session_id")
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
		sessionID := c.Param("game_session_id")
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

func ConnectToGameSession(rdb *db.RedisClient, upgrader websocket.Upgrader) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionID := c.Param("game_session_id")
		if sessionID == "" {
			log.Println("Error: Game session ID is missing")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Game session ID is required"})
		}
		log.Printf("Attempting to connect to game session: %s", sessionID)

		// Get user information from the context
		userID := c.Get("userID").(string)
		username := c.Get("username").(string)
		log.Printf("User connecting: ID=%s, Username=%s", userID, username)

		// Get the current game session
		gameSession, err := rdb.GetGameSession(c.Request().Context(), uuid.MustParse(sessionID))
		if err != nil {
			log.Printf("Error getting game session: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get game session"})
		}
		log.Printf("Retrieved game session: %+v", gameSession)

		// Check if the user is already in the game session
		userExists := false
		for _, player := range gameSession.Players {
			if player.ID.String() == userID {
				userExists = true
				break
			}
		}
		log.Printf("User exists in session: %v", userExists)

		// If the user is not in the game session, add them
		if !userExists {
			newPlayer := models.User{
				ID:       uuid.MustParse(userID),
				Username: username,
			}
			gameSession.Players = append(gameSession.Players, newPlayer)

			// Update the game session in Redis
			err = rdb.UpdateGameSession(c.Request().Context(), gameSession)
			if err != nil {
				log.Printf("Error updating game session: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update game session"})
			}
			log.Println("Added new player to game session")
		}

		// Upgrade the HTTP connection to a WebSocket connection
		log.Println("Attempting to upgrade to WebSocket connection")
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("Error upgrading to WebSocket: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to upgrade to WebSocket"})
		}
		defer func() {
			ws.Close()
			log.Printf("Closing WebSocket connection for user %s", userID)
			removePlayerFromSession(c.Request().Context(), rdb, uuid.MustParse(sessionID), uuid.MustParse(userID))
		}()

		ctx, cancel := context.WithCancel(c.Request().Context())
		defer cancel()

		updates, err := rdb.SubscribeToGameSession(ctx, uuid.MustParse(sessionID))
		if err != nil {
			log.Printf("Error subscribing to game session: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to subscribe to game session"})
		}

		// Send initial game session data
		if err := ws.WriteJSON(gameSession); err != nil {
			log.Printf("Error sending initial game session data: %v", err)
			return nil
		}

		// Channel for WebSocket read errors
		readErr := make(chan error)

		// Start a goroutine to read from WebSocket
		go func() {
			for {
				messageType, message, err := ws.ReadMessage()
				if err != nil {
					log.Printf("Error reading message: %v", err)
					readErr <- err
					return
				}
				log.Printf("Received message from client (type %d): %s", messageType, string(message))
			}
		}()

		// Main event loop
		for {
			select {
			case update := <-updates:
				if err := ws.WriteJSON(update); err != nil {
					log.Printf("Error sending update to client: %v", err)
					return nil
				}
			case err := <-readErr:
				log.Printf("WebSocket read error: %v", err)
				return nil
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func removePlayerFromSession(ctx context.Context, rdb *db.RedisClient, sessionID, userID uuid.UUID) {
	gameSession, err := rdb.GetGameSession(ctx, sessionID)
	if err != nil {
		// Log the error, but don't return it as this is a cleanup function
		log.Printf("Failed to get game session: %v", err)
		return
	}

	// Remove the player from the game session
	for i, player := range gameSession.Players {
		if player.ID == userID {
			gameSession.Players = append(gameSession.Players[:i], gameSession.Players[i+1:]...)
			break
		}
	}

	// Update the game session in Redis
	err = rdb.UpdateGameSession(ctx, gameSession)
	if err != nil {
		log.Printf("Failed to update game session: %v", err)
	}
}

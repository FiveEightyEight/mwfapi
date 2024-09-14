package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type Score struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Points   int8      `json:"points"`
}

type Game struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Users  []User    `json:"users"`
	Scores []Score   `json:"score"`
}

type GameSession struct {
	ID        uuid.UUID `json:"id"`
	GameID    uuid.UUID `json:"game_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Players   []User    `json:"players"`
	Scores    []Score   `json:"scores"`
	Status    string    `json:"status"`
}

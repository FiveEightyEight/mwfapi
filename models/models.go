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

type GameSessionStatus string

const (
	GameSessionStatusWaiting    GameSessionStatus = "waiting"
	GameSessionStatusInProgress GameSessionStatus = "in_progress"
	GameSessionStatusFinished   GameSessionStatus = "finished"
)

type GameSession struct {
	ID                  uuid.UUID         `json:"id"`
	Name                string            `json:"name"`
	GameID              uuid.UUID         `json:"game_id"`
	GameConfig          GameConfig        `json:"game_config"`
	Problems            []GameProblem     `json:"problems"`
	CurrentProblemIndex int               `json:"current_problem_index"`
	StartTime           time.Time         `json:"start_time"`
	EndTime             time.Time         `json:"end_time"`
	Players             []User            `json:"players"`
	Scores              []Score           `json:"scores"`
	Status              GameSessionStatus `json:"status"`
}

type ActiveGameSession struct {
	Games []uuid.UUID `json:"games"`
}

type GameConfigRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type GameConfigMethod string

const (
	GameConfigMethodAdd      GameConfigMethod = "add"
	GameConfigMethodSubtract GameConfigMethod = "subtract"
	GameConfigMethodMultiply GameConfigMethod = "multiply"
	GameConfigMethodDivide   GameConfigMethod = "divide"
)

type GameConfig struct {
	Methods []GameConfigMethod `json:"methods"`
	Range   GameConfigRange    `json:"range"`
}

type GameProblem struct {
	Number1 int              `json:"number1"`
	Number2 int              `json:"number2"`
	Method  GameConfigMethod `json:"method"`
	Answer  int              `json:"answer"`
}

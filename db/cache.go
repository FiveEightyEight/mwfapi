package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FiveEightyEight/mwfapi/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(opts *redis.Options) *RedisClient {
	client := redis.NewClient(opts)
	return &RedisClient{client: client}
}

// User CRUD operations

func (rc *RedisClient) CreateUser(ctx context.Context, user *models.User) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return rc.client.Set(ctx, fmt.Sprintf("user:%s", user.ID), userJSON, 0).Err()
}

func (rc *RedisClient) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	userJSON, err := rc.client.Get(ctx, fmt.Sprintf("user:%s", id)).Bytes()
	if err != nil {
		return nil, err
	}
	var user models.User
	err = json.Unmarshal(userJSON, &user)
	return &user, err
}

func (rc *RedisClient) UpdateUser(ctx context.Context, user *models.User) error {
	return rc.CreateUser(ctx, user) // Same as create since we're overwriting
}

func (rc *RedisClient) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return rc.client.Del(ctx, fmt.Sprintf("user:%s", id)).Err()
}

// Score CRUD operations

func (rc *RedisClient) CreateScore(ctx context.Context, score *models.Score) error {
	scoreJSON, err := json.Marshal(score)
	if err != nil {
		return err
	}
	return rc.client.Set(ctx, fmt.Sprintf("score:%s", score.ID), scoreJSON, 0).Err()
}

func (rc *RedisClient) GetScore(ctx context.Context, id uuid.UUID) (*models.Score, error) {
	scoreJSON, err := rc.client.Get(ctx, fmt.Sprintf("score:%s", id)).Bytes()
	if err != nil {
		return nil, err
	}
	var score models.Score
	err = json.Unmarshal(scoreJSON, &score)
	return &score, err
}

func (rc *RedisClient) UpdateScore(ctx context.Context, score *models.Score) error {
	return rc.CreateScore(ctx, score) // Same as create since we're overwriting
}

func (rc *RedisClient) DeleteScore(ctx context.Context, id uuid.UUID) error {
	return rc.client.Del(ctx, fmt.Sprintf("score:%s", id)).Err()
}

// Game CRUD operations

func (rc *RedisClient) CreateGame(ctx context.Context, game *models.Game) error {
	gameJSON, err := json.Marshal(game)
	if err != nil {
		return err
	}
	return rc.client.Set(ctx, fmt.Sprintf("game:%s", game.ID), gameJSON, 0).Err()
}

func (rc *RedisClient) GetGame(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	gameJSON, err := rc.client.Get(ctx, fmt.Sprintf("game:%s", id)).Bytes()
	if err != nil {
		return nil, err
	}
	var game models.Game
	err = json.Unmarshal(gameJSON, &game)
	return &game, err
}

func (rc *RedisClient) UpdateGame(ctx context.Context, game *models.Game) error {
	return rc.CreateGame(ctx, game) // Same as create since we're overwriting
}

func (rc *RedisClient) DeleteGame(ctx context.Context, id uuid.UUID) error {
	return rc.client.Del(ctx, fmt.Sprintf("game:%s", id)).Err()
}

// GameSession CRUD operations

func (rc *RedisClient) CreateGameSession(ctx context.Context, gameSession *models.GameSession) error {
	gameSessionJSON, err := json.Marshal(gameSession)
	if err != nil {
		return err
	}
	return rc.client.Set(ctx, fmt.Sprintf("game_session:%s", gameSession.ID), gameSessionJSON, 0).Err()
}

func (rc *RedisClient) GetGameSession(ctx context.Context, id uuid.UUID) (*models.GameSession, error) {
	gameSessionJSON, err := rc.client.Get(ctx, fmt.Sprintf("game_session:%s", id)).Bytes()
	if err != nil {
		return nil, err
	}
	var gameSession models.GameSession
	err = json.Unmarshal(gameSessionJSON, &gameSession)
	return &gameSession, err
}

func (rc *RedisClient) UpdateGameSession(ctx context.Context, gameSession *models.GameSession) error {
	return rc.CreateGameSession(ctx, gameSession) // Same as create since we're overwriting
}

func (rc *RedisClient) DeleteGameSession(ctx context.Context, id uuid.UUID) error {
	return rc.client.Del(ctx, fmt.Sprintf("game_session:%s", id)).Err()
}

// Subscribe to GameSession

func (rc *RedisClient) SubscribeToGameSession(ctx context.Context, gameSessionID uuid.UUID) (<-chan *models.GameSession, error) {
	pubsub := rc.client.Subscribe(ctx, fmt.Sprintf("game_session:%s", gameSessionID))

	ch := make(chan *models.GameSession)

	go func() {
		defer pubsub.Close()
		defer close(ch)

		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				return
			}

			var gameSession models.GameSession
			if err := json.Unmarshal([]byte(msg.Payload), &gameSession); err != nil {
				continue
			}

			ch <- &gameSession
		}
	}()

	return ch, nil
}

// Publish GameSession update

func (rc *RedisClient) PublishGameSessionUpdate(ctx context.Context, gameSession *models.GameSession) error {
	gameSessionJSON, err := json.Marshal(gameSession)
	if err != nil {
		return err
	}
	return rc.client.Publish(ctx, fmt.Sprintf("game_session:%s", gameSession.ID), gameSessionJSON).Err()
}

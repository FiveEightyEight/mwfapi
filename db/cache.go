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
	err = rc.client.Set(ctx, fmt.Sprintf("game_session:%s", gameSession.ID), gameSessionJSON, 0).Err()
	if err != nil {
		return err
	}
	return rc.UpdateActiveGameSessions(ctx, gameSession.ID, true)
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
	err := rc.CreateGameSession(ctx, gameSession) // Same as create since we're overwriting
	if err != nil {
		return err
	}

	// Publish the updated game session
	return rc.PublishGameSessionUpdate(ctx, gameSession)
}

func (rc *RedisClient) DeleteGameSession(ctx context.Context, id uuid.UUID) error {
	err := rc.client.Del(ctx, fmt.Sprintf("game_session:%s", id)).Err()
	if err != nil {
		return err
	}
	return rc.UpdateActiveGameSessions(ctx, id, false)
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

func (rc *RedisClient) GetActiveGameSessions(ctx context.Context) ([]*models.GameSession, error) {
	key := "active_game_sessions"
	sessionIDs, err := rc.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var activeSessions []*models.GameSession
	for _, idStr := range sessionIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		session, err := rc.GetGameSession(ctx, id)
		if err != nil {
			continue
		}
		activeSessions = append(activeSessions, session)
	}

	return activeSessions, nil
}

func (rc *RedisClient) CloseGameSession(ctx context.Context, sessionID uuid.UUID) error {
	err := rc.DeleteGameSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return rc.UpdateActiveGameSessions(ctx, sessionID, false)
}

func (rc *RedisClient) UpdateActiveGameSessions(ctx context.Context, gameSessionID uuid.UUID, add bool) error {
	key := "active_game_sessions"

	if add {
		return rc.client.SAdd(ctx, key, gameSessionID.String()).Err()
	} else {
		return rc.client.SRem(ctx, key, gameSessionID.String()).Err()
	}
}

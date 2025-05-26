package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"golepi-fsm/internal/domain/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client        *redis.Client
	keyPrefix     string
	defaultExpiry time.Duration
}

func NewRedisStorage(config redis.Options, keyPrefix string, defaultExpiry time.Duration) *RedisStorage {
	if keyPrefix == "" {
		keyPrefix = "fsm:"
	}

	if defaultExpiry == 0 {
		defaultExpiry = 24 * time.Hour
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return &RedisStorage{
		client:        client,
		keyPrefix:     keyPrefix,
		defaultExpiry: defaultExpiry,
	}
}

func (s *RedisStorage) makeStateKey(chatID, userID int64) string {
	return fmt.Sprintf("%sstate:%d:%d", s.keyPrefix, chatID, userID)
}

func (s *RedisStorage) makeDataKey(chatID, userID int64) string {
	return fmt.Sprintf("%sdata:%d:%d", s.keyPrefix, chatID, userID)
}

func (s *RedisStorage) Get(ctx context.Context, chatID int64, userID int64) (models.State, error) {
	key := s.makeStateKey(chatID, userID)
	val, err := s.client.Get(ctx, key).Result()

	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to get state from Redis: %w", err)
	}

	return models.State(val), nil
}

func (s *RedisStorage) Set(ctx context.Context, chatID int64, userID int64, state models.State) error {
	key := s.makeStateKey(chatID, userID)
	err := s.client.Set(ctx, key, string(state), s.defaultExpiry).Err()
	if err != nil {
		return fmt.Errorf("failed to set state in Redis: %w", err)
	}
	return nil
}

func (s *RedisStorage) Delete(ctx context.Context, chatID int64, userID int64) error {
	key := s.makeStateKey(chatID, userID)
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete state from Redis: %w", err)
	}
	return nil
}

func (s *RedisStorage) GetData(ctx context.Context, chatID int64, userID int64, key string) (interface{}, error) {
	dataKey := s.makeDataKey(chatID, userID)
	val, err := s.client.HGet(ctx, dataKey, key).Result()

	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get data from Redis: %w", err)
	}

	var result any
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data from Redis: %w", err)
	}

	return result, nil
}

func (s *RedisStorage) SetData(ctx context.Context, chatID int64, userID int64, key string, value any) error {
	dataKey := s.makeDataKey(chatID, userID)

	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data for Redis: %w", err)
	}

	if err := s.client.HSet(ctx, dataKey, key, jsonData).Err(); err != nil {
		return fmt.Errorf("failed to set data in Redis: %w", err)
	}

	if err := s.client.Expire(ctx, dataKey, s.defaultExpiry).Err(); err != nil {
		return fmt.Errorf("failed to set expiry for Redis key: %w", err)
	}

	return nil
}

func (s *RedisStorage) ClearData(ctx context.Context, chatID int64, userID int64) error {
	dataKey := s.makeDataKey(chatID, userID)
	err := s.client.Del(ctx, dataKey).Err()
	if err != nil {
		return fmt.Errorf("failed to clear data from Redis: %w", err)
	}
	return nil
}

func (s *RedisStorage) Close() error {
	return s.client.Close()
}

func (s *RedisStorage) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

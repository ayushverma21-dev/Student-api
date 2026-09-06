package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ayushverma21-dev/Student-api/internal/config"
	"github.com/ayushverma21-dev/Student-api/internal/types"
	redisclient "github.com/redis/go-redis/v9"
)

const sessionTTL = 24 * time.Hour

type Store struct {
	client *redisclient.Client
}

func New(cfg config.Redis) (*Store, error) {
	client := redisclient.NewClient(&redisclient.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	return &Store{client: client}, nil
}

func (store *Store) Close() error {
	return store.client.Close()
}

func (store *Store) CreateSession(ctx context.Context, student types.Student) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	value, err := json.Marshal(student)
	if err != nil {
		return "", fmt.Errorf("encode session: %w", err)
	}

	if err := store.client.Set(ctx, "session:"+token, value, sessionTTL).Err(); err != nil {
		return "", fmt.Errorf("store session: %w", err)
	}
	return token, nil
}

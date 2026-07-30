package cache

import (
	"context"
	"encoding/json"
	"login/internal/config"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedis(cfg *config.RedisConfig, log *zap.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Info("Redis 连接成功", zap.String("addr", cfg.Addr()))
	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

// 黑名单相关

const BlacklistPrefix = "blk:"

func (r *RedisClient) AddToBlacklist(ctx context.Context, jti string, ttl time.Duration) error {
	return r.client.Set(ctx, BlacklistPrefix+jti, time.Now().Unix(), ttl).Err()
}

func (r *RedisClient) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := r.client.Exists(ctx, BlacklistPrefix+jti).Result()
	return n > 0, err
}

// 会话管理

const SessionPrefix = "sessions:"

type SessionData struct {
	RefreshJTI   string `json:"refresh_jti"`
	UserAgent    string `json:"user_agent"`
	IP           string `json:"ip"`
	CreatedAt    int64  `json:"created_at"`
	LastAccessAt int64  `json:"last_access_at"`
}

func (r *RedisClient) SaveSession(ctx context.Context, userID uint64, deviceID string, data *SessionData) error {
	return r.client.HSet(ctx, SessionPrefix+formatUint64(userID), deviceID, marshal(data)).Err()
}

func (r *RedisClient) GetSession(ctx context.Context, userID uint64, deviceID string) (*SessionData, error) {
	data, err := r.client.HGet(ctx, SessionPrefix+formatUint64(userID), deviceID).Result()
	if err != nil {
		return nil, err
	}
	return unmarshalSession(data)
}

func (r *RedisClient) GetAllSessions(ctx context.Context, userID uint64) (map[string]*SessionData, error) {
	result, err := r.client.HGetAll(ctx, SessionPrefix+formatUint64(userID)).Result()
	if err != nil {
		return nil, err
	}
	sessions := make(map[string]*SessionData)
	for deviceID, data := range result {
		s, err := unmarshalSession(data)
		if err != nil {
			continue
		}
		sessions[deviceID] = s
	}
	return sessions, nil
}

func (r *RedisClient) DeleteSession(ctx context.Context, userID uint64, deviceID string) error {
	return r.client.HDel(ctx, SessionPrefix+formatUint64(userID), deviceID).Err()
}

func (r *RedisClient) DeleteAllSessions(ctx context.Context, userID uint64) error {
	return r.client.Del(ctx, SessionPrefix+formatUint64(userID)).Err()
}

// 限流

const RateLimitPrefix = "rate_limit:"

func (r *RedisClient) IncrementRateLimit(ctx context.Context, key string, ttl time.Duration) (int, error) {
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, RateLimitPrefix+key)
	pipe.Expire(ctx, RateLimitPrefix+key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return int(incr.Val()), nil
}

func (r *RedisClient) ResetRateLimit(ctx context.Context, key string) error {
	return r.client.Del(ctx, RateLimitPrefix+key).Err()
}

// helper
func formatUint64(n uint64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func marshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func unmarshalSession(data string) (*SessionData, error) {
	var s SessionData
	err := json.Unmarshal([]byte(data), &s)
	return &s, err
}


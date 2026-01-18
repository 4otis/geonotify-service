package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrNotFound = redis.Nil
)

type Client struct {
	client *redis.Client
	ctx    context.Context
}

func NewClient(ctx context.Context, url string) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}, nil
}

func (c *Client) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.Set(c.ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

func (c *Client) Get(key string, dest interface{}) error {
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err == redis.Nil {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

func (c *Client) Delete(key string) error {
	if err := c.client.Del(c.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

func (c *Client) LPush(queue string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.LPush(c.ctx, queue, data).Err(); err != nil {
		return fmt.Errorf("failed to push to queue %s: %w", queue, err)
	}

	return nil
}

func (c *Client) BRPop(queue string, timeout time.Duration) (string, []byte, error) {
	result, err := c.client.BRPop(c.ctx, timeout, queue).Result()
	if err == redis.Nil {
		return "", nil, ErrNotFound
	}
	if err != nil {
		return "", nil, fmt.Errorf("failed to BRPop from queue %s: %w", queue, err)
	}

	if len(result) != 2 {
		return "", nil, fmt.Errorf("invalid BRPop result: %v", result)
	}

	return result[0], []byte(result[1]), nil
}

func (c *Client) ZAdd(queue string, score float64, member interface{}) error {
	data, err := json.Marshal(member)
	if err != nil {
		return fmt.Errorf("failed to marshal member: %w", err)
	}

	if err := c.client.ZAdd(c.ctx, queue, &redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("failed to ZAdd to queue %s: %w", queue, err)
	}

	return nil
}

func (c *Client) ZRangeByScore(queue string, min, max string, offset, count int64) ([][]byte, error) {
	opt := &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}

	members, err := c.client.ZRangeByScore(c.ctx, queue, opt).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ZRangeByScore from queue %s: %w", queue, err)
	}

	result := make([][]byte, len(members))
	for i, member := range members {
		result[i] = []byte(member)
	}

	return result, nil
}

func (c *Client) ZRem(queue string, member interface{}) error {
	data, err := json.Marshal(member)
	if err != nil {
		return fmt.Errorf("failed to marshal member: %w", err)
	}

	if err := c.client.ZRem(c.ctx, queue, data).Err(); err != nil {
		return fmt.Errorf("failed to ZRem from queue %s: %w", queue, err)
	}

	return nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) HealthCheck() error {
	return c.client.Ping(c.ctx).Err()
}

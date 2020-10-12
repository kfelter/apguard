package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// DB ...
type DB interface {
	Set(key string, value interface{}, ttl time.Duration)
	Get(key string) (interface{}, error)
}

// RedisClient ...
type RedisClient struct {
	standaloneClient *redis.Client
	clusterClient    *redis.ClusterClient
}

// NewRC ...
func NewRC(url string) (*RedisClient, error) {
	sp := strings.Split(url, ":")
	if len(sp) < 2 {
		return nil, fmt.Errorf("malformed redis url: %s", url)
	}
	if sp[0] == "redis" {
		opts, err := redis.ParseURL(url)
		if err != nil {
			return nil, fmt.Errorf("malformed redis url opts: %s: %v", url, err)
		}
		c := redis.NewClient(opts)
		if err := c.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("could not connect to redis url: %s: %v", url, err)
		}
		return &RedisClient{
			standaloneClient: c,
		}, nil
	}

	if sp[0] == "cluster" {
		c := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{"redis:" + sp[1]},
		})

		if err := c.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("could not connect to %s: %v", url, err)
		}

		return &RedisClient{
			clusterClient: c,
		}, nil
	}

	return nil, fmt.Errorf("not implemented %s", url)
}

// Set ...
func (rc *RedisClient) Set(key string, value interface{}, ttl time.Duration) {
	if rc.standaloneClient != nil {
		rc.standaloneClient.Set(context.Background(), key, value, ttl)
	}
	if rc.clusterClient != nil {
		rc.clusterClient.Set(context.Background(), key, value, ttl)
	}
}

// Get ...
func (rc *RedisClient) Get(key string) (interface{}, error) {
	if rc.standaloneClient != nil {
		return rc.standaloneClient.Get(context.Background(), key).Result()
	}
	if rc.clusterClient != nil {
		return rc.clusterClient.Get(context.Background(), key).Result()
	}
	return nil, fmt.Errorf("rcs are nil")
}

package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type DB redis.Client

func NewDB(redisURL string) (*DB, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	db := redis.NewClient(opt)
	if err := db.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return (*DB)(db), nil
}

func (db *DB) LoadSessions() map[string]string {
	rooms, err := db.Keys(context.Background(), "*").Result()
	if err != nil {
		log.Println("Redis error:", err)
		return nil
	}
	results := make(map[string]string)
	for _, room := range rooms {
		fen, err := db.Get(context.Background(), room).Result()
		if err != nil {
			log.Println("Redis error:", err)
			continue
		}
		results[room] = fen
	}
	return results
}

func (db *DB) SaveSession(room, fen string, expiration time.Duration) {
	if err := db.Set(context.Background(), room, fen, expiration).Err(); err != nil {
		log.Println("Redis error:", err)
	}
}

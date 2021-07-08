package redis

import (
	"log"
	"time"

	"github.com/go-redis/redis/v7"
)

func connection() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := rdb.Ping().Result()
	if err != nil {
		log.Fatal(err.Error())
	}
	return rdb, nil
}

func SetKey(key string, email string, expires time.Duration) error {
	rdb, err := connection()
	defer rdb.Close()
	if err != nil {
		return err
	}
	err = rdb.Set(key, email, expires).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetKey(str string) (string, error) {
	rdb, err := connection()
	defer rdb.Close()
	if err != nil {
		return "", err
	}
	key, err := rdb.Get(str).Result()
	if err != nil {
		return "", err
	}
	return key, nil
}

func Deletekey(str string) (int64, error) {
	rdb, err := connection()
	defer rdb.Close()
	if err != nil {
		return 0, err
	}
	deleted, err := rdb.Del(str).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

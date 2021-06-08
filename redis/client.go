package redis

import (
	"github.com/go-redis/redis/v7"
)

var client *redis.Client

func Init() {
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", //redis port
		Password: "",
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
}

func GetClient() *redis.Client {
	return client
}

// Redis는 오픈 소스 인메모리 데이터 구조 저장소
// https://jjeongil.tistory.com/1403

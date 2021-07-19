package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	URL  string `json:"url"`
	User string `json:"username"`
	Pwd  string `json:"password"`
}

// opt : server => main plc => edge
func getAuth() (Config, error) {
	var auth Config
	conf, err := os.Open("dbConfig.json")
	if err != nil {
		return auth, err
	}
	byteVal, _ := ioutil.ReadAll(conf)
	json.Unmarshal(byteVal, &auth)

	return auth, nil
}

func getConnection() (*mongo.Client, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conf, err := getAuth()
	if err != nil {
		return nil, fmt.Errorf("DB Err : %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.URL).SetAuth((options.Credential{
		Username: conf.User,
		Password: conf.Pwd,
	})))

	if err != nil {
		return nil, fmt.Errorf("DB Err : %v", err)
	}

	return client, nil
}

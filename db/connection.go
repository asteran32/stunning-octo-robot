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

type DBConfig struct {
	Url string `json:url`
	Db  string `json:db`
	Col string `json:collection`
}

var UserClient *mongo.Client
var UserCollection *mongo.Collection

// opt : server => main plc => edge
func DbInit() (DBConfig, error) {
	var dbCof DBConfig

	conf, err := os.Open("dbConfig.json")
	if err != nil {
		return dbCof, err
	}
	byteVal, _ := ioutil.ReadAll(conf)
	json.Unmarshal(byteVal, &dbCof)
	return dbCof, nil
}

func getConnection() error {
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	conf, err := DbInit()
	if err != nil {
		return fmt.Errorf("DB Err : %v", err)
	}
	UserClient, err = mongo.Connect(ctx, options.Client().ApplyURI(conf.Url))
	if err != nil {
		return fmt.Errorf("DB Err : %v", err)
	}
	UserCollection = UserClient.Database("testApp").Collection("user")

	return nil
}

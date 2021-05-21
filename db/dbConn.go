package db

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	Server []DBConfig `json:"server"`
}

type PLC struct {
	Plc []DBConfig `json:"plc"`
}

type DBConfig struct {
	URL        string `json:url`
	Username   string `json:username`
	Password   string `json:password`
	Database   string `json:database`
	Collection string `json:collection`
}

var UserClient *mongo.Client
var UserCollection *mongo.Collection

// opt : server => main plc => edge
func getDBConfig() (DBConfig, error) {
	var dbCof DBConfig
	conf, err := os.Open("dbConfig.json")
	if err != nil {
		return dbCof, err
	}
	byteVal, _ := ioutil.ReadAll(conf)
	json.Unmarshal(byteVal, &dbCof)
	return dbCof, nil
}

func dbConnect() {
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	conf, err := getDBConfig()
	if err != nil {
		log.Fatal("DB Err:", err)
	}

	UserClient, err = mongo.Connect(ctx, options.Client().ApplyURI(conf.URL).SetAuth((options.Credential{
		Username: conf.Username,
		Password: conf.Password,
	})))

	if err != nil {
		log.Fatal("DB Err:", err)
	}
	UserCollection = UserClient.Database(conf.Database).Collection(conf.Collection)

	log.Println("Success to Connect MonogDB")
}

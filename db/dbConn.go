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

type DBConfig struct {
	URL        string `json:url`
	Username   string `json:username`
	Password   string `json:password`
	Database   string `json:database`
	Collection string `json:collection`
}

var UserClient *mongo.Client
var UserCollection *mongo.Collection

var PlcClient *mongo.Client
var PlcCollection *mongo.Collection

// opt : server => main plc => edge
func getDBConfig(opt string) (DBConfig, error) {
	var dbCof DBConfig
	var fname string
	switch opt {
	case "plc":
		fname = "opcDBconfig.json"
	case "user":
		fname = "dbConfig.json"
	}
	conf, err := os.Open(fname)
	if err != nil {
		return dbCof, err
	}
	byteVal, _ := ioutil.ReadAll(conf)
	json.Unmarshal(byteVal, &dbCof)
	return dbCof, nil
}

func dbConnect(opt string) {
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	conf, err := getDBConfig(opt)
	if err != nil {
		log.Fatal("DB Err:", err)
	}

	switch opt {
	case "plc":
		PlcClient, err = mongo.Connect(ctx, options.Client().ApplyURI(conf.URL))
		if err != nil {
			log.Fatal("DB Err:", err)
		}
		PlcCollection = PlcClient.Database(conf.Database).Collection(conf.Collection)

	case "user":
		UserClient, err = mongo.Connect(ctx, options.Client().ApplyURI(conf.URL).SetAuth((options.Credential{
			Username: conf.Username,
			Password: conf.Password,
		})))
		if err != nil {
			log.Fatal("DB Err:", err)
		}
		UserCollection = UserClient.Database(conf.Database).Collection(conf.Collection)

	}
}

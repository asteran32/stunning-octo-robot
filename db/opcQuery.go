package db

import (
	"app/model"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func GetCSVList() (model.Csvs, error) {
	if PlcClient == nil {
		dbConnect("plc")
	}
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	csvlist := model.Csvs{}
	// err := PlcCollection.FindOne(ctx, bson.M{"date": bson.M{"$gt": "2021-05-23"}}).Decode(&csvList)
	cur, err := PlcCollection.Find(ctx, bson.M{})
	if err != nil {
		return csvlist, err
	}
	// Decoding
	for cur.Next(ctx) {
		var csv model.Csv
		e := cur.Decode(&csv)
		if e != nil {
			log.Fatal(e)
		}
		csvlist.Items = append(csvlist.Items, csv)
	}

	return csvlist, nil
}

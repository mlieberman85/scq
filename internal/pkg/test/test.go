package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
	b "gopkg.in/mgo.v2/bson"
)

func UploadTestData(testDataDir string) error {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost"))
	if err != nil {
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	scDatabase := client.Database(("supplychain"))
	err = scDatabase.Collection("attestations").Drop(ctx)
	if err != nil {
		return err
	}
	attestationCollection := scDatabase.Collection("attestations")
	mod := mongo.IndexModel{
		Keys: bson.M{
			"subject.digest.sha256": 1,
		},
		Options: nil,
	}

	_, err = attestationCollection.Indexes().CreateOne(ctx, mod)
	if err != nil {
		return err
	}

	var testAttestations []interface{}
	files, err := ioutil.ReadDir(testDataDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			data, err := os.ReadFile(fmt.Sprintf("%s/%s", testDataDir, file.Name()))
			if err != nil {
				return err
			}

			var bdoc bson.M
			err = b.UnmarshalJSON(data, &bdoc)
			if err != nil {
				return err
			}

			testAttestations = append(testAttestations, bdoc)
		}
	}
	attestationCollection.InsertMany(ctx, testAttestations)

	return nil
}

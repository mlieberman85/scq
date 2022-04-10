package record

import (
	"context"
	"encoding/json"

	"github.com/in-toto/in-toto-golang/in_toto"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type DocDBRecord struct {
	Attestation Attestation
}

func (d *DocDBRecord) GetData() interface{} {
	return d.Attestation
}

func (d *DocDBRecord) GetMaterials() []Material {
	var ms []Material
	for _, predicateMaterial := range d.Attestation.(*in_toto.ProvenanceStatement).Predicate.Materials {
		// For now just get the first hash in the map.
		var d string
		for _, digest := range predicateMaterial.Digest {
			d = digest
			break
		}

		m := Material{
			MediaType: "todo",
			Digest:    d,
			Uri:       predicateMaterial.URI,
		}
		ms = append(ms, m)
	}

	return ms
}

func (d *DocDBRecord) GetType() string {
	return "docdbrecord"
}

type MongoClient struct {
	ctx        context.Context
	collection mongo.Collection
	client     *mongo.Client
}

func GetMongoClient(uri string, db string, collection string) (*MongoClient, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri)) // mongodb://localhost
	if err != nil {
		return nil, err
	}

	err = client.Connect(context.Background())
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	scDatabase := client.Database((db))                         // supplychain
	attestationsCollection := scDatabase.Collection(collection) // attestations

	return &MongoClient{
		ctx:        context.Background(),
		collection: *attestationsCollection,
		client:     client,
	}, nil
}

func (mc *MongoClient) GetRecord(hash string) (Record, error) {
	c, err := mc.collection.Find(mc.ctx, bson.M{"subject.digest.sha256": hash}, options.Find().SetProjection(bson.M{"_id": 0}))
	if err != nil {
		return nil, err
	}

	var results []bson.M
	err = c.All(mc.ctx, &results)
	if err != nil {
		return nil, err
	}

	r := DocDBRecord{
		Attestation: nil,
	}

	if len(results) > 0 {
		d := results[0]
		m, err := bson.Marshal(d)
		if err != nil {
			return nil, err
		}

		var p *in_toto.ProvenanceStatement
		var p2 interface{}

		// FIXME: For some reason bson doesn't unmarshal the statementheader correctly
		// But json marshalling/unmarshalling does
		/*err = bson.Unmarshal(m, &p)
		if err != nil {
			return nil, err
		}*/
		err = bson.Unmarshal(m, &p2)
		if err != nil {
			return nil, err
		}
		j, err := json.Marshal(p2)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(j, &p)
		r.Attestation = p
	}

	return &r, nil
}

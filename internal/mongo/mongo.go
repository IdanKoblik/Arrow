package mongo

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"arrow/pkg/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Collection   *mongo.Collection
	historyCache struct {
		sync.RWMutex
		data []types.AlertJSON
	}
)

func Init() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	db := os.Getenv("MONGODB_DB")
	if db == "" {
		db = "alerts_db"
	}

	Collection = client.Database(db).Collection("alerts")

	Collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	Collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "received_at", Value: -1}},
	})

	log.Println("[MongoDB] connected")
}

func StoreAlert(p *types.OrefPayload) {
	doc := types.AlertDoc{
		OrefID:     p.ID,
		Cat:        p.Cat,
		Title:      p.Title,
		Data:       p.Data,
		Desc:       p.Desc,
		ReceivedAt: time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Collection.InsertOne(ctx, doc)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		log.Println(err)
	}

	loc, _ := time.LoadLocation("Asia/Jerusalem")
	alertJSON := types.AlertJSON{
		ID:        doc.OrefID,
		Cat:       doc.Cat,
		Title:     doc.Title,
		Data:      doc.Data,
		Desc:      doc.Desc,
		AlertDate: doc.ReceivedAt.In(loc).Format("2006-01-02 15:04:05"),
	}

	historyCache.Lock()
	historyCache.data = append([]types.AlertJSON{alertJSON}, historyCache.data...)
	if len(historyCache.data) > 200 {
		historyCache.data = historyCache.data[:200]
	}
	historyCache.Unlock()
}

func GetHistory() []types.AlertJSON {
	historyCache.RLock()
	if len(historyCache.data) > 0 {
		data := historyCache.data
		historyCache.RUnlock()
		return data
	}
	historyCache.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := Collection.Find(ctx, bson.D{},
		options.Find().
			SetSort(bson.D{{Key: "received_at", Value: -1}}).
			SetLimit(200),
	)
	if err != nil {
		return []types.AlertJSON{}
	}
	defer cur.Close(ctx)

	var docs []types.AlertDoc
	if err := cur.All(ctx, &docs); err != nil {
		return []types.AlertJSON{}
	}

	loc, _ := time.LoadLocation("Asia/Jerusalem")

	out := make([]types.AlertJSON, len(docs))
	for i, d := range docs {
		out[i] = types.AlertJSON{
			ID:    d.OrefID,
			Cat:   d.Cat,
			Title: d.Title,
			Data:  d.Data,
			Desc:  d.Desc,
			AlertDate: d.ReceivedAt.In(loc).
				Format("2006-01-02 15:04:05"),
		}
	}

	historyCache.Lock()
	historyCache.data = out
	historyCache.Unlock()

	return out
}

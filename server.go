package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	orefURL = "https://www.oref.org.il/WarningMessages/alert/alerts.json"
)

type AlertDoc struct {
	OrefID     string    `bson:"id"`
	Cat        string    `bson:"cat"`
	Title      string    `bson:"title"`
	Data       []string  `bson:"data"`
	Desc       string    `bson:"desc"`
	ReceivedAt time.Time `bson:"received_at"`
}

type AlertJSON struct {
	ID        string   `json:"id"`
	Cat       string   `json:"cat"`
	Title     string   `json:"title"`
	Data      []string `json:"data"`
	Desc      string   `json:"desc"`
	AlertDate string   `json:"alertDate"`
}

type orefPayload struct {
	ID    string   `json:"id"`
	Cat   string   `json:"cat"`
	Title string   `json:"title"`
	Data  []string `json:"data"`
	Desc  string   `json:"desc"`
}

var (
	col        *mongo.Collection
	httpClient = &http.Client{
		Timeout: 4 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    5,
			IdleConnTimeout: 90 * time.Second,
		},
	}
)

func initMongo() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("MongoDB connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping: %v", err)
	}

	col = client.Database("alerts_db").Collection("alerts")

	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "received_at", Value: -1}},
	})

	log.Printf("[MongoDB] connected — %s", uri)
}

func storeAlert(p *orefPayload) {
	doc := AlertDoc{
		OrefID:     p.ID,
		Cat:        p.Cat,
		Title:      p.Title,
		Data:       p.Data,
		Desc:       p.Desc,
		ReceivedAt: time.Now(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := col.InsertOne(ctx, doc)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		log.Printf("[MongoDB] insert: %v", err)
	}
}

func getHistory() []AlertJSON {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := col.Find(ctx, bson.D{},
		options.Find().
			SetSort(bson.D{{Key: "received_at", Value: -1}}).
			SetLimit(200),
	)
	if err != nil {
		log.Printf("[MongoDB] find: %v", err)
		return []AlertJSON{}
	}
	defer cur.Close(ctx)

	var docs []AlertDoc
	if err := cur.All(ctx, &docs); err != nil {
		log.Printf("[MongoDB] decode: %v", err)
		return []AlertJSON{}
	}

	out := make([]AlertJSON, len(docs))
	for i, d := range docs {
		out[i] = AlertJSON{
			ID:        d.OrefID,
			Cat:       d.Cat,
			Title:     d.Title,
			Data:      d.Data,
			Desc:      d.Desc,
			AlertDate: d.ReceivedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return out
}

func jsonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store")
}

func handleAlerts(w http.ResponseWriter, r *http.Request) {
	req, _ := http.NewRequest("GET", orefURL, nil)
	req.Header.Set("Referer", "https://www.oref.org.il/")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	trimmed := strings.TrimSpace(strings.TrimPrefix(string(body), "\xef\xbb\xbf"))
	if trimmed != "" && trimmed != "null" && trimmed != "\r\n" {
		var p orefPayload
		if json.Unmarshal([]byte(trimmed), &p) == nil && p.ID != "" {
			storeAlert(&p)
		}
	}

	jsonHeaders(w)
	w.Write(body)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	jsonHeaders(w)
	json.NewEncoder(w).Encode(getHistory())
}

func main() {
	log.Println("Starting server...")
	initMongo()

	fs := http.FileServer(http.Dir("."))
	mux := http.NewServeMux()
	mux.HandleFunc("/api/alerts", handleAlerts)
	mux.HandleFunc("/api/history", handleHistory)
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}
		fs.ServeHTTP(w, r)
	}))

	addr := os.Getenv("ADDR_INFO")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}

	log.Printf("Serving at http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

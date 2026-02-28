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
	col         *mongo.Collection
	latestAlert *orefPayload
	lastID      string

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

	log.Println("[MongoDB] connected")
}

func storeAlert(p *orefPayload) {
	doc := AlertDoc{
		OrefID:     p.ID,
		Cat:        p.Cat,
		Title:      p.Title,
		Data:       p.Data,
		Desc:       p.Desc,
		ReceivedAt: time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := col.InsertOne(ctx, doc)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		log.Println("[MongoDB] insert:", err)
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
		return []AlertJSON{}
	}
	defer cur.Close(ctx)

	var docs []AlertDoc
	if err := cur.All(ctx, &docs); err != nil {
		return []AlertJSON{}
	}

	loc, _ := time.LoadLocation("Asia/Jerusalem")

	out := make([]AlertJSON, len(docs))
	for i, d := range docs {
		out[i] = AlertJSON{
			ID:        d.OrefID,
			Cat:       d.Cat,
			Title:     d.Title,
			Data:      d.Data,
			Desc:      d.Desc,
			AlertDate: d.ReceivedAt.In(loc).
				Format("2006-01-02 15:04:05"),
		}
	}
	return out
}

func pollOref() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		req, _ := http.NewRequest("GET", orefURL, nil)
		req.Header.Set("Referer", "https://www.oref.org.il/")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Accept", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Println("[Oref] request error:", err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		trimmed := strings.TrimSpace(strings.TrimPrefix(string(body), "\xef\xbb\xbf"))
		if trimmed == "" || trimmed == "null" || trimmed == "\r\n" {
			continue
		}

		var p orefPayload
		if err := json.Unmarshal([]byte(trimmed), &p); err != nil {
			continue
		}

		if p.ID == "" {
			continue
		}

		if p.ID != lastID {
			log.Println("[Oref] New alert:", p.ID)
			lastID = p.ID
			latestAlert = &p
			storeAlert(&p)
		}
	}
}

func jsonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
}

func handleAlerts(w http.ResponseWriter, r *http.Request) {
	jsonHeaders(w)

	if latestAlert == nil {
		w.Write([]byte("null"))
		return
	}

	json.NewEncoder(w).Encode(latestAlert)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	jsonHeaders(w)
	json.NewEncoder(w).Encode(getHistory())
}

func main() {
	log.Println("Starting server...")

	initMongo()

	go pollOref()

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
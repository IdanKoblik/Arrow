package types

import "time"

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

type OrefPayload struct {
	ID    string   `json:"id"`
	Cat   string   `json:"cat"`
	Title string   `json:"title"`
	Data  []string `json:"data"`
	Desc  string   `json:"desc"`
}
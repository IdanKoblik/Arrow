package api

import (
	"encoding/json"
	"net/http"

	"arrow/internal/mongo"
	"arrow/internal/oref"
)

func jsonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
}

func HandleAlerts(w http.ResponseWriter, r *http.Request) {
	jsonHeaders(w)

	if oref.LatestAlert == nil {
		w.Write([]byte("null"))
		return
	}

	json.NewEncoder(w).Encode(oref.LatestAlert)
}

func HandleHistory(w http.ResponseWriter, r *http.Request) {
	jsonHeaders(w)
	json.NewEncoder(w).Encode(mongo.GetHistory())
}
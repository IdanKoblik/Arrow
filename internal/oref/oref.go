package oref

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"arrow/internal/mongo"
	"arrow/pkg/types"
)

const orefURL = "https://www.oref.org.il/WarningMessages/alert/alerts.json"

var (
	LatestAlert *types.OrefPayload
	lastID      string
)

var httpClient = &http.Client{
	Timeout: 4 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:    5,
		IdleConnTimeout: 90 * time.Second,
	},
}

func StartPolling() {
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
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		trimmed := strings.TrimSpace(strings.TrimPrefix(string(body), "\xef\xbb\xbf"))
		if trimmed == "" || trimmed == "null" {
			continue
		}

		var p types.OrefPayload
		if err := json.Unmarshal([]byte(trimmed), &p); err != nil {
			continue
		}

		if p.ID == "" {
			continue
		}

		if p.ID != lastID {
			log.Println("[Oref] New alert:", p.ID)
			lastID = p.ID
			LatestAlert = &p
			mongo.StoreAlert(&p)
		}
	}
}
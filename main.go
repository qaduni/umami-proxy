package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type UmamiStat struct {
	Pageviews struct {
		Value int `json:"value"`
	} `json:"pageviews"`
	Visitors struct {
		Value int `json:"value"`
	} `json:"visitors"`
}

type PeriodStats struct {
	Pageviews int `json:"pageviews"`
	Visitors  int `json:"visitors"`
}

type ProxyResponse struct {
	WebsiteID string      `json:"website_id"`
	Today     PeriodStats `json:"today"`
	Week      PeriodStats `json:"week"`
	Month     PeriodStats `json:"month"`
	Total     PeriodStats `json:"total"`
}

var umamiURL string

// Shared HTTP Client with TCP connection pooling across all request threads
var sharedClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}

func fetchStat(websiteID string, startAt, endAt int64) (PeriodStats, error) {
	url := fmt.Sprintf("%s/api/websites/%s/stats?startAt=%d&endAt=%d", umamiURL, websiteID, startAt, endAt)

	resp, err := sharedClient.Get(url)
	if err != nil {
		return PeriodStats{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PeriodStats{}, fmt.Errorf("umami returned status %d", resp.StatusCode)
	}

	var stat UmamiStat
	if err := json.NewDecoder(resp.Body).Decode(&stat); err != nil {
		return PeriodStats{}, err
	}

	return PeriodStats{
		Pageviews: stat.Pageviews.Value,
		Visitors:  stat.Visitors.Value,
	}, nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	websiteID := r.URL.Query().Get("id")
	if websiteID == "" {
		http.Error(w, `{"error":"missing 'id' query parameter"}`, http.StatusBadRequest)
		return
	}

	now := time.Now()
	nowMs := now.UnixMilli()

	todayStartMs := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()

	offset := int(now.Weekday()) - 1
	if offset < 0 {
		offset = 6
	}
	weekStartMs := time.Date(now.Year(), now.Month(), now.Day()-offset, 0, 0, 0, 0, now.Location()).UnixMilli()
	monthStartMs := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).UnixMilli()

	var wg sync.WaitGroup
	var response ProxyResponse
	response.WebsiteID = websiteID

	var errToday, errWeek, errMonth, errTotal error

	wg.Add(4)
	go func() { defer wg.Done(); response.Today, errToday = fetchStat(websiteID, todayStartMs, nowMs) }()
	go func() { defer wg.Done(); response.Week, errWeek = fetchStat(websiteID, weekStartMs, nowMs) }()
	go func() { defer wg.Done(); response.Month, errMonth = fetchStat(websiteID, monthStartMs, nowMs) }()
	go func() { defer wg.Done(); response.Total, errTotal = fetchStat(websiteID, 0, nowMs) }()
	wg.Wait()

	if errToday != nil || errWeek != nil || errMonth != nil || errTotal != nil {
		http.Error(w, `{"error":"failed to fetch statistics from umami"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	umamiURL = os.Getenv("UMAMI_URL")
	if umamiURL == "" {
		log.Fatal("FATAL: UMAMI_URL environment variable is required")
	}

	http.HandleFunc("/stats", statsHandler)
	log.Printf("Proxy running on port 8080 forwarding to %s...", umamiURL)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

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

// Cache entry struct
type cacheItem struct {
	data      ProxyResponse
	expiresAt time.Time
}

// MemoryCache handles thread-safe in-memory caching
type MemoryCache struct {
	sync.RWMutex
	items map[string]cacheItem
	ttl   time.Duration
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	c := &MemoryCache{
		items: make(map[string]cacheItem),
		ttl:   ttl,
	}
	// Background worker to purge expired keys periodically
	go func() {
		for range time.Tick(2 * ttl) {
			c.Lock()
			now := time.Now()
			for k, item := range c.items {
				if now.After(item.expiresAt) {
					delete(c.items, k)
				}
			}
			c.Unlock()
		}
	}()
	return c
}

func (c *MemoryCache) Get(key string) (ProxyResponse, bool) {
	c.RLock()
	defer c.RUnlock()
	item, found := c.items[key]
	if !found || time.Now().After(item.expiresAt) {
		return ProxyResponse{}, false
	}
	return item.data, true
}

func (c *MemoryCache) Set(key string, data ProxyResponse) {
	c.Lock()
	defer c.Unlock()
	c.items[key] = cacheItem{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
}

var (
	umamiURL    string
	umamiAPIKey string
	statsCache  *MemoryCache
)

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

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return PeriodStats{}, fmt.Errorf("failed to create request: %w", err)
	}

	if umamiAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+umamiAPIKey)
	}

	resp, err := sharedClient.Do(req)
	if err != nil {
		return PeriodStats{}, fmt.Errorf("network error calling Umami: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PeriodStats{}, fmt.Errorf("umami HTTP %d from endpoint %s", resp.StatusCode, url)
	}

	var stat UmamiStat
	if err := json.NewDecoder(resp.Body).Decode(&stat); err != nil {
		return PeriodStats{}, fmt.Errorf("failed to decode json: %w", err)
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

	// 1. Check Cache First
	if cachedData, found := statsCache.Get(websiteID); found {
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(cachedData)
		return
	}
	w.Header().Set("X-Cache", "MISS")

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

	// 2. Comprehensive Logging & Error Response
	if errToday != nil || errWeek != nil || errMonth != nil || errTotal != nil {
		log.Printf("[ERROR] Website %s fetch failed:\n  Today: %v\n  Week: %v\n  Month: %v\n  Total: %v",
			websiteID, errToday, errWeek, errMonth, errTotal)

		http.Error(w, `{"error":"failed to fetch statistics from umami"}`, http.StatusInternalServerError)
		return
	}

	// 3. Save to Cache
	statsCache.Set(websiteID, response)

	json.NewEncoder(w).Encode(response)
}

func main() {
	umamiURL = os.Getenv("UMAMI_URL")
	if umamiURL == "" {
		log.Fatal("FATAL: UMAMI_URL environment variable is required")
	}

	umamiAPIKey = os.Getenv("UMAMI_API_KEY")

	// Initialize 5-minute in-memory cache
	statsCache = NewMemoryCache(5 * time.Minute)

	http.HandleFunc("/stats", statsHandler)
	log.Printf("Proxy running on port 8080 forwarding to %s...", umamiURL)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

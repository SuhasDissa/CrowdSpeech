package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "crowdspeech-backend/migrations"
)

// ── Rate Limiter ─────────────────────────────────────────────────────────────

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *rateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	var valid []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		return false
	}

	rl.requests[ip] = append(valid, now)
	return true
}

func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for ip, reqs := range rl.requests {
			var valid []time.Time
			for _, t := range reqs {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.SplitN(forwarded, ",", 2)[0])
	}
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return real
	}
	return r.RemoteAddr
}

// ── Daily batch scheduler ─────────────────────────────────────────────────────

// scheduleDailyProcessing fires at midnight UTC every day.
func scheduleDailyProcessing(app *pocketbase.PocketBase) {
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
		time.Sleep(time.Until(next))
		log.Println("batch processor: starting daily run")
		processBatch(app, false)
	}
}

// ── Main ─────────────────────────────────────────────────────────────────────

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	rl := newRateLimiter(20, time.Minute)

	// Rate-limit recording creation
	app.OnRecordCreateRequest("recordings").BindFunc(func(e *core.RecordRequestEvent) error {
		ip := getClientIP(e.Request)
		if !rl.Allow(ip) {
			return apis.NewTooManyRequestsError("Rate limit exceeded. Please wait before submitting more recordings.", nil)
		}
		return e.Next()
	})

	// Update keyword counts after each new recording
	app.OnRecordAfterCreateSuccess("recordings").BindFunc(func(e *core.RecordEvent) error {
		go updateKeywordCount(app, e.Record)
		return e.Next()
	})

	// Register custom routes
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/api/export", handleExport(app))
		se.Router.GET("/api/stats", handleStats(app))
		se.Router.POST("/api/seed", handleSeed(app))
		se.Router.POST("/api/process", handleProcess(app))
		return se.Next()
	})

	// Seed + start batch scheduler after PocketBase is ready
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		go func() {
			time.Sleep(2 * time.Second)
			if err := seedIfEmpty(app); err != nil {
				log.Printf("seed error: %v", err)
			}
			go scheduleDailyProcessing(app)
		}()
		return se.Next()
	})

	port := os.Getenv("PB_PORT")
	if port == "" {
		port = "10001"
	}
	dataDir := os.Getenv("PB_DATA_DIR")
	if dataDir == "" {
		dataDir = "./pb_data"
	}

	app.RootCmd.SetArgs([]string{
		"serve",
		"--http=0.0.0.0:" + port,
		"--dir=" + dataDir,
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const groceriesURL = "https://gist.githubusercontent.com/SuhasDissa/a8fb0e736c4b3b494eff3f6bbc846fa7/raw/5987579469a0a830abd231621c0cd95a167bf7fd/groceries.json"

type Translation struct {
	Scripts         []string `json:"scripts"`
	Transliterations []string `json:"transliterations"`
}

type GroceryItem struct {
	ItemName         string                 `json:"itemName"`
	AlternativeNames []string               `json:"alternativeNames"`
	Category         string                 `json:"category"`
	Translations     map[string]Translation `json:"translations"`
}

func fetchGroceries() ([]GroceryItem, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(groceriesURL)
	if err != nil {
		return nil, fmt.Errorf("fetch groceries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groceries fetch status: %d", resp.StatusCode)
	}

	var items []GroceryItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode groceries: %w", err)
	}
	return items, nil
}

func seedIfEmpty(app *pocketbase.PocketBase) error {
	col, err := app.FindCollectionByNameOrId("keywords")
	if err != nil {
		return fmt.Errorf("find keywords collection: %w", err)
	}

	existing, err := app.FindAllRecords("keywords")
	if err != nil {
		return fmt.Errorf("count keywords: %w", err)
	}
	if len(existing) > 0 {
		log.Printf("Keywords already seeded (%d records), skipping.", len(existing))
		return nil
	}

	log.Println("Seeding keywords from groceries JSON...")
	items, err := fetchGroceries()
	if err != nil {
		return fmt.Errorf("fetch groceries for seed: %w", err)
	}

	const sampleTarget = 100
	created := 0

	for _, item := range items {
		if item.ItemName == "" {
			continue
		}

		siScript := firstOf(item.Translations["sinhala"].Scripts)
		taScript := firstOf(item.Translations["tamil"].Scripts)

		record := core.NewRecord(col)
		record.Set("text", item.ItemName)   // canonical English label
		record.Set("text_si", siScript)
		record.Set("text_ta", taScript)
		record.Set("category", item.Category)
		record.Set("sample_target", sampleTarget)
		record.Set("count_en", 0)
		record.Set("count_si", 0)
		record.Set("count_ta", 0)

		if err := app.Save(record); err != nil {
			log.Printf("Failed to save keyword %q: %v", item.ItemName, err)
			continue
		}
		created++
	}

	log.Printf("Seed complete: %d keywords created.", created)
	return nil
}

func firstOf(ss []string) string {
	if len(ss) > 0 {
		return ss[0]
	}
	return ""
}

// handleSeed is an admin endpoint to re-run seeding (idempotent)
func handleSeed(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if !isAdminAuthed(e) {
			return e.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		go func() {
			if err := seedIfEmpty(app); err != nil {
				log.Printf("Manual seed error: %v", err)
			}
		}()
		return e.JSON(http.StatusAccepted, map[string]string{"status": "seeding started"})
	}
}

func updateKeywordCount(app *pocketbase.PocketBase, rec *core.Record) {
	kwID := rec.GetString("keyword")
	lang := rec.GetString("language")
	if kwID == "" || lang == "" {
		return
	}

	kw, err := app.FindRecordById("keywords", kwID)
	if err != nil {
		log.Printf("updateKeywordCount: find keyword %s: %v", kwID, err)
		return
	}

	// Count recordings for this keyword + language combination
	records, err := app.FindAllRecords("recordings",
		dbx.HashExp{"keyword": kwID, "language": lang},
	)
	if err != nil {
		log.Printf("updateKeywordCount: count recordings: %v", err)
		return
	}

	countField := "count_" + lang
	kw.Set(countField, len(records))
	if err := app.Save(kw); err != nil {
		log.Printf("updateKeywordCount: save: %v", err)
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// batchRunning prevents overlapping batch runs (daily + manual).
var batchRunning atomic.Bool

// processBatch converts all unprocessed recordings (audio_wav == "") to WAV.
// Safe to call concurrently — only one run executes at a time.
func processBatch(app *pocketbase.PocketBase) {
	if !batchRunning.CompareAndSwap(false, true) {
		log.Println("batch processor: already running, skipping")
		return
	}
	defer batchRunning.Store(false)

	records, err := app.FindAllRecords("recordings")
	if err != nil {
		log.Printf("batch processor: fetch recordings: %v", err)
		return
	}

	pending := make([]*core.Record, 0)
	for _, r := range records {
		if r.GetString("audio_wav") == "" && r.GetString("audio") != "" {
			pending = append(pending, r)
		}
	}

	if len(pending) == 0 {
		log.Println("batch processor: nothing to process")
		return
	}

	log.Printf("batch processor: processing %d recordings", len(pending))

	// Process with limited concurrency (4 parallel ffmpeg jobs)
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup

	for _, rec := range pending {
		wg.Add(1)
		sem <- struct{}{}
		go func(r *core.Record) {
			defer wg.Done()
			defer func() { <-sem }()
			processAudio(app, r)
		}(rec)
	}

	wg.Wait()
	log.Println("batch processor: done")
}

// processAudio converts a single recording to 16kHz 16-bit mono WAV.
func processAudio(app *pocketbase.PocketBase, rec *core.Record) {
	audioFilename := rec.GetString("audio")
	if audioFilename == "" {
		return
	}

	srcPath := filepath.Join(
		app.DataDir(), "storage",
		rec.Collection().Id,
		rec.Id,
		audioFilename,
	)

	base := strings.TrimSuffix(audioFilename, filepath.Ext(audioFilename))
	wavFilename := base + ".wav"
	wavPath := filepath.Join(
		app.DataDir(), "storage",
		rec.Collection().Id,
		rec.Id,
		wavFilename,
	)

	if err := convertToWAV(srcPath, wavPath); err != nil {
		log.Printf("processAudio: convert %s: %v", rec.Id, err)
		return
	}

	// Re-fetch to avoid overwriting concurrent field updates
	fresh, err := app.FindRecordById("recordings", rec.Id)
	if err != nil {
		log.Printf("processAudio: refetch %s: %v", rec.Id, err)
		os.Remove(wavPath)
		return
	}

	fresh.Set("audio_wav", wavFilename)
	if err := app.Save(fresh); err != nil {
		log.Printf("processAudio: save %s: %v", rec.Id, err)
		os.Remove(wavPath)
	} else {
		log.Printf("processAudio: processed %s → %s", rec.Id, wavFilename)
	}
}

// convertToWAV converts any audio to 16kHz 16-bit mono WAV with EBU R128
// normalization and silence trimming.
func convertToWAV(src, dst string) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source file %q not found: %w", src, err)
	}

	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", src,
		"-af", strings.Join([]string{
			"silenceremove=start_periods=1:start_silence=0.1:start_threshold=-50dB" +
				":stop_periods=1:stop_silence=0.5:stop_threshold=-50dB",
			"loudnorm=I=-16:TP=-1.5:LRA=11",
		}, ","),
		"-ar", "16000",
		"-ac", "1",
		"-c:a", "pcm_s16le",
		dst,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(out))
	}
	return nil
}

// handleProcess is an admin endpoint that triggers the batch processor immediately.
func handleProcess(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if !isAdminAuthed(e) {
			return e.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		if batchRunning.Load() {
			return e.JSON(http.StatusConflict, map[string]string{"status": "already running"})
		}
		go processBatch(app)
		return e.JSON(http.StatusAccepted, map[string]string{"status": "batch processing started"})
	}
}

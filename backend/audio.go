package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// processAudio runs ffmpeg to convert uploaded audio to 16kHz 16-bit mono WAV,
// normalize audio, and trim leading/trailing silence.
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

	// Re-fetch the record so we have the latest version
	fresh, err := app.FindRecordById("recordings", rec.Id)
	if err != nil {
		log.Printf("processAudio: refetch %s: %v", rec.Id, err)
		os.Remove(wavPath)
		return
	}

	fresh.Set("audio_wav", wavFilename)
	if err := app.Save(fresh); err != nil {
		log.Printf("processAudio: save wav ref for %s: %v", rec.Id, err)
		os.Remove(wavPath)
	} else {
		log.Printf("processAudio: processed %s → %s", rec.Id, wavFilename)
	}
}

// convertToWAV converts any audio to 16kHz 16-bit mono WAV with normalization and silence trimming.
func convertToWAV(src, dst string) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}

	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source file %q not found: %w", src, err)
	}

	// silenceremove: trim leading/trailing silence at -50dB threshold
	// loudnorm: EBU R128 loudness normalization to -16 LUFS
	// -ar 16000: 16kHz sample rate
	// -ac 1: mono
	// -c:a pcm_s16le: 16-bit signed PCM WAV
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

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

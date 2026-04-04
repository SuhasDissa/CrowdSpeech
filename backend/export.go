package main

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/crypto/bcrypt"
)

func isAdminAuthed(e *core.RequestEvent) bool {
	hash := os.Getenv("ADMIN_PASSWORD_HASH")
	if hash == "" {
		return false
	}
	password := e.Request.Header.Get("X-Admin-Password")
	if password == "" {
		password = e.Request.URL.Query().Get("password")
	}
	if password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func handleExport(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if !isAdminAuthed(e) {
			e.Response.Header().Set("WWW-Authenticate", `Basic realm="CrowdSpeech Admin"`)
			return e.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized. Provide X-Admin-Password header or ?password= query param.",
			})
		}

		lang := e.Request.URL.Query().Get("language")

		var recordings []*core.Record
		var err error
		if lang != "" {
			recordings, err = app.FindAllRecords("recordings",
				dbx.HashExp{"language": lang},
			)
		} else {
			recordings, err = app.FindAllRecords("recordings")
		}
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		e.Response.Header().Set("Content-Type", "application/zip")
		e.Response.Header().Set("Content-Disposition", `attachment; filename="crowdspeech_export.zip"`)

		zw := zip.NewWriter(e.Response)
		defer zw.Close()

		// CSV metadata writer
		csvFile, err := zw.Create("metadata.csv")
		if err != nil {
			return err
		}
		cw := csv.NewWriter(csvFile)
		_ = cw.Write([]string{
			"id", "language", "keyword_id", "keyword_en", "keyword_display",
			"audio_file", "duration", "created",
			"age_group", "gender", "country", "primary_language", "accent",
			"region", "education", "years_speaking", "occupation", "speech_condition",
		})

		for _, rec := range recordings {
			kwID := rec.GetString("keyword")
			recLang := rec.GetString("language")
			kwEn := ""
			kwDisplay := ""
			if kw, kerr := app.FindRecordById("keywords", kwID); kerr == nil {
				kwEn = kw.GetString("text")
				switch recLang {
				case "si":
					kwDisplay = kw.GetString("text_si")
				case "ta":
					kwDisplay = kw.GetString("text_ta")
				default:
					kwDisplay = kwEn
				}
			}

			// audio_wav is a TextField (plain filename set by batch processor).
			// audio is a FileField — use GetStringSlice and take the first entry.
			audioFilename := rec.GetString("audio_wav")
			if audioFilename == "" {
				if files := rec.GetStringSlice("audio"); len(files) > 0 {
					audioFilename = files[0]
				}
			}

			// Always write the CSV row; mark audio as pending if not yet available.
			zipEntryName := ""
			if audioFilename != "" {
				audioPath := filepath.Join(
					app.DataDir(), "storage",
					rec.Collection().Id,
					rec.Id,
					audioFilename,
				)
				zipEntryName = fmt.Sprintf("audio/%s/%s_%s", recLang, rec.Id, audioFilename)
				if err := addFileToZip(zw, audioPath, zipEntryName); err != nil {
					zipEntryName = "(file missing: " + audioFilename + ")"
				}
			} else {
				zipEntryName = "(pending ffmpeg processing)"
			}

			_ = cw.Write([]string{
				rec.Id,
				recLang,
				kwID,
				kwEn,
				kwDisplay,
				zipEntryName,
				fmt.Sprintf("%.2f", rec.GetFloat("duration")),
				rec.GetDateTime("created").String(),
				rec.GetString("age_group"),
				rec.GetString("gender"),
				rec.GetString("country"),
				rec.GetString("primary_language"),
				rec.GetString("accent"),
				rec.GetString("region"),
				rec.GetString("education"),
				rec.GetString("years_speaking"),
				rec.GetString("occupation"),
				rec.GetString("speech_condition"),
			})
		}

		cw.Flush()
		return nil
	}
}

func addFileToZip(zw *zip.Writer, srcPath, destName string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = destName
	header.Method = zip.Deflate

	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func handleStats(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		type LangStats struct {
			Language   string `json:"language"`
			Recordings int    `json:"recordings"`
		}

		langs := []string{"en", "si", "ta"}
		stats := make([]LangStats, 0, len(langs))

		for _, lang := range langs {
			recs, _ := app.FindAllRecords("recordings", dbx.HashExp{"language": lang})
			stats = append(stats, LangStats{Language: lang, Recordings: len(recs)})
		}

		total, _ := app.FindAllRecords("recordings")

		return e.JSON(http.StatusOK, map[string]any{
			"total_recordings": len(total),
			"by_language":      stats,
		})
	}
}

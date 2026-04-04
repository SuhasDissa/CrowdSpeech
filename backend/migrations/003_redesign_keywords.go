package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Redesign keywords: one record per item (labeled by English itemName),
// with separate display-text fields per language and per-language counts.
func init() {
	m.Register(func(app core.App) error {
		// Drop old collections (recordings first — it holds a relation to keywords)
		for _, name := range []string{"recordings", "keywords"} {
			col, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue // already absent
			}
			if err := app.Delete(col); err != nil {
				return err
			}
		}

		// ── keywords (one record per grocery item) ───────────────────────────
		keywords := core.NewBaseCollection("keywords")
		keywords.ListRule = types.Pointer("") // public read
		keywords.ViewRule = types.Pointer("") // public read
		keywords.Fields.Add(
			// Canonical English label used in dataset exports
			&core.TextField{Name: "text", Required: true},
			// Per-language display text shown to the speaker
			&core.TextField{Name: "text_si"}, // Sinhala script
			&core.TextField{Name: "text_ta"}, // Tamil script
			&core.TextField{Name: "category"},
			&core.NumberField{Name: "sample_target", Min: floatPtr(0)},
			// Per-language recording counts (for balanced selection)
			&core.NumberField{Name: "count_en", Min: floatPtr(0)},
			&core.NumberField{Name: "count_si", Min: floatPtr(0)},
			&core.NumberField{Name: "count_ta", Min: floatPtr(0)},
		)

		if err := app.Save(keywords); err != nil {
			return err
		}

		// ── recordings ───────────────────────────────────────────────────────
		recordings := core.NewBaseCollection("recordings")
		recordings.CreateRule = types.Pointer("") // public create
		recordings.Fields.Add(
			&core.TextField{Name: "language", Required: true},
			&core.RelationField{
				Name:         "keyword",
				CollectionId: keywords.Id,
				Required:     true,
				MaxSelect:    1,
			},
			&core.FileField{
				Name:      "audio",
				MaxSelect: 1,
				MaxSize:   52428800,
				MimeTypes: []string{
					"audio/webm", "audio/ogg", "audio/wav",
					"audio/mp4", "audio/mpeg",
				},
			},
			// Filename only — WAV is written directly to storage by ffmpeg hook,
		// bypassing PocketBase's upload mechanism.
		&core.TextField{Name: "audio_wav"},
			&core.NumberField{Name: "duration", Min: floatPtr(0)},
			&core.BoolField{Name: "validated"},
			&core.TextField{Name: "client_ip"},
		)

		return app.Save(recordings)
	}, nil) // no rollback — previous migrations handle the base state
}

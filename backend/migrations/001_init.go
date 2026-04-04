package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		// ── keywords collection ──────────────────────────────────────────────
		keywords := core.NewBaseCollection("keywords")
		keywords.ListRule = types.Pointer("") // public read
		keywords.ViewRule = types.Pointer("") // public read
		keywords.Fields.Add(
			&core.TextField{Name: "language", Required: true},
			&core.TextField{Name: "text", Required: true},
			&core.TextField{Name: "category"},
			&core.NumberField{Name: "sample_target", Min: floatPtr(0)},
			&core.NumberField{Name: "current_count", Min: floatPtr(0)},
		)

		if err := app.Save(keywords); err != nil {
			return err
		}

		// ── recordings collection ────────────────────────────────────────────
		recordings := core.NewBaseCollection("recordings")
		recordings.CreateRule = types.Pointer("") // public create (anonymous submissions)
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
				MaxSize:   52428800, // 50 MB
				MimeTypes: []string{
					"audio/webm",
					"audio/ogg",
					"audio/wav",
					"audio/mp4",
					"audio/mpeg",
				},
			},
			&core.FileField{
				Name:      "audio_wav",
				MaxSelect: 1,
				MaxSize:   52428800,
				MimeTypes: []string{"audio/wav", "audio/x-wav"},
			},
			&core.NumberField{Name: "duration", Min: floatPtr(0)},
			&core.BoolField{Name: "validated"},
			&core.TextField{Name: "client_ip"},
		)

		return app.Save(recordings)
	}, func(app core.App) error {
		for _, name := range []string{"recordings", "keywords"} {
			col, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue
			}
			if err := app.Delete(col); err != nil {
				return err
			}
		}
		return nil
	})
}

func floatPtr(f float64) *float64 { return &f }

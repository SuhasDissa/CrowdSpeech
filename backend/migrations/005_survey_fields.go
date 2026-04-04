package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Add demographic survey fields to recordings.
// All stored as plain text (the selected option label).
func init() {
	m.Register(func(app core.App) error {
		recordings, err := app.FindCollectionByNameOrId("recordings")
		if err != nil {
			return err
		}

		recordings.Fields.Add(
			&core.TextField{Name: "age_group"},
			&core.TextField{Name: "gender"},
			&core.TextField{Name: "country"},
			&core.TextField{Name: "primary_language"},
			&core.TextField{Name: "accent"},
			&core.TextField{Name: "region"},
			&core.TextField{Name: "education"},
			&core.TextField{Name: "years_speaking"},
			&core.TextField{Name: "occupation"},
			&core.TextField{Name: "speech_condition"},
		)

		return app.Save(recordings)
	}, nil)
}

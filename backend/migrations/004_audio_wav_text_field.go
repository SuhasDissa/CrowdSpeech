package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Change audio_wav from FileField to TextField so the ffmpeg hook can write
// a plain filename string without going through PocketBase's upload mechanism.
func init() {
	m.Register(func(app core.App) error {
		recordings, err := app.FindCollectionByNameOrId("recordings")
		if err != nil {
			return err
		}

		existing := recordings.Fields.GetByName("audio_wav")
		if existing != nil {
			recordings.Fields.RemoveById(existing.GetId())
		}

		recordings.Fields.Add(&core.TextField{Name: "audio_wav"})

		return app.Save(recordings)
	}, nil)
}

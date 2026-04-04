package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		// keywords: allow public list + view
		keywords, err := app.FindCollectionByNameOrId("keywords")
		if err != nil {
			return err
		}
		keywords.ListRule = types.Pointer("")
		keywords.ViewRule = types.Pointer("")
		if err := app.Save(keywords); err != nil {
			return err
		}

		// recordings: allow public create (anonymous submissions)
		recordings, err := app.FindCollectionByNameOrId("recordings")
		if err != nil {
			return err
		}
		recordings.CreateRule = types.Pointer("")
		return app.Save(recordings)
	}, func(app core.App) error {
		// rollback: restore admin-only rules
		for _, name := range []string{"keywords", "recordings"} {
			col, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue
			}
			col.ListRule = nil
			col.ViewRule = nil
			col.CreateRule = nil
			_ = app.Save(col)
		}
		return nil
	})
}

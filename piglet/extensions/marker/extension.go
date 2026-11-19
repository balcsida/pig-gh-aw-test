// Package marker is a minimal extension, scaffolded with `pig extension init`.
// It exists so that the gh-aw-ext Piglet has a non-empty extensions: section.
package marker

import (
	sdk "github.com/MichaelKinsy/PiG/extensions/sdk"
)

func Extension() *sdk.Extension {
	e := sdk.New("marker")

	e.Tool("marker_ping", "Return a short pong so you can confirm the extension is wired up.",
		sdk.Schema{"type": "object", "properties": map[string]any{}},
		func(ctx sdk.Context, params map[string]any) (any, error) {
			return "pong from marker", nil
		})

	return e
}

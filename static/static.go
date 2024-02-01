package static

import "embed"

// EmbeddedStatic holds static web server content.
//   - "js/", "views/" are always served from embed.FS.
//   - "css/", "font/" and "img/" are only served from
//     embed.FS when no (or no existing) -style-dir has been provided.
//
//go:embed *
var EmbeddedStatic embed.FS

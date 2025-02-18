package static

import "embed"

// EmbeddedStatic holds static web server content.
// This is the original content of gnoweb (the first version, prior to Q4 2024)
//
//go:embed *
var EmbeddedStatic embed.FS

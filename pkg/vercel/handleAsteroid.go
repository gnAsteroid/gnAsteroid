package vercel

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/grepsuzette/gnAsteroid"
)

// Run the whole thing (asteroid+gnoweb) as an http.Handler.
// This function is exported so you can use this on Vercel.
//
// In that case, no server is launched & Vercel will run its own.
//
// Both asteroid and style are expected as fs.FS.
// style == nil means the same as gnAsteroid.DefaultStyle().
//
// CAUTION: Remember (as you will be using embed),
// to first fs.Sub(yourEmbed, "subDir"), because //go:embed
// won't do that automatically. Otherwise files in your
// asteroid won't be found, e.g. it will look in "index.md"
// instead of "myAsteroid/index.md" and Vercel will give
// you a 404. Same thing for style if you don't use DefaultStyle().
//
// See the test file for an example.
//
func HandleAsteroid(asteroid, style fs.FS, asteroidName_ string, cfg gnoweb.Config) http.Handler {
	gnAsteroid.SetAsteroidFs(asteroid)
	gnAsteroid.SetAsteroidName(asteroidName_)
	return gnAsteroid.MakeApp(slog.Default(), cfg, style)
}

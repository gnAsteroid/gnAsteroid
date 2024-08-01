package gnAsteroid

import (
	"embed"
	"html"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gotuna/gotuna"
	gnosmos "github.com/grepsuzette/gnosmos.style"
	"github.com/yalue/merged_fs"
)

// Instead of serving a realm-based editorial blog
// about gno.land, this uses gnoweb to serve markdown
// based websites with customizable style. The other
// features offered by gnoweb (rendering r/demo/foo)
// otherwise remain.
//
// In other words, website thus create still gravitate
// around the gno.land chain (hence, "asteroids", if
// gno.land is a plane(t)).

var (
	asteroidName string // read from cmdLine, or file called .TITLE at root, or "CHANGEME"
	asteroidFs   fs.FS  // ultimately used by HandlerHome and HandlerAnything
)

//go:embed views/*.html
var newViews embed.FS // composed with gnoweb's, using merged_fs

func SetAsteroidFs(asteroid fs.FS) { asteroidFs = asteroid }
func SetAsteroidName(name string)  { asteroidName = name }
func DefaultStyle() fs.FS {
	return gnosmos.Style
}

// Encapsulate a serverless gnoweb serving an asteroid as an http.Handler
// It can then be served, for example on Vercel. nil style uses gnosmos.Style
func HandleAsteroid(asteroid, style fs.FS, asteroidName_ string, cfg gnoweb.Config) http.Handler {
	SetAsteroidFs(asteroid)
	SetAsteroidName(asteroidName_)
	return MakeApp(slog.Default(), cfg, style)
}

// And this is separate from HandleAsteroid because?...
// ...mainly because cmd/main uses MakeApp() to reload
// watched files.
func MakeApp(logger *slog.Logger, cfg gnoweb.Config, styleFs fs.FS) http.Handler {
	gnowebViews, e := fs.Sub(gnoweb.DefaultViewsFiles(), "views")
	if e != nil {
		panic("Could not find gnoweb views: " + e.Error())
	}
	asteroidViews, e := fs.Sub(newViews, "views")
	if e != nil {
		panic("Could not find asteroid views: " + e.Error())
	}
	if styleFs == nil {
		styleFs = gnosmos.Style
	}
	return gnoweb.MakeAppWithOptions(logger, cfg, gnoweb.Options{
		RootHandler:     HandlerRoot,
		NotFoundHandler: HandlerAnything,
		StyleFS:         styleFs,
		ViewFS:          merged_fs.NewMergedFS(asteroidViews, gnowebViews),
	}).Router
}

// handler serving /
// it is used when gnoweb needs to serve a root index.md or README.md file.
func HandlerRoot(logger *slog.Logger, app gotuna.App, cfg *gnoweb.Config) http.Handler {
	var rootFile fs.File
	var e error
	rootFile, e = asteroidFs.Open("index.md")
	if e != nil {
		rootFile, e = asteroidFs.Open("README.md")
	}
	if e != nil {
		panic("asteroid must include /(index|README).md")
	}
	fileInfo, _ := rootFile.Stat()
	buf := make([]byte, fileInfo.Size())

	_, e = rootFile.Read(buf)
	if e != nil {
		panic(e.Error())
	}
	// homeContent := osm.MustReadFile(file)
	homeContent := buf
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Set("AsteroidName", asteroidName).
			Set("HomeContent", string(homeContent)).
			Set("Config", cfg).
			Render(w, r, "asteroidHome.html", "funcs.html")
	})
}

// this is a fallthrough handler to
// attempt serving markdown, images
// in the structure.
func HandlerAnything(logger *slog.Logger, app gotuna.App, cfg *gnoweb.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		url = html.UnescapeString(url)
		url = strings.TrimPrefix(url, "/") // need to remove leading / with embed fs
		url = filepath.Clean(url)

		if strings.Contains(url, "..") {
			app.NewTemplatingEngine().
				Set("AsteroidName", asteroidName).
				Set("Config", cfg).
				Render(w, r, "asteroid403.html", "funcs.html")
		}
		var file fs.File // nil means still unfound
		servedFilename := ""

		for _, path := range []string{
			url + "/index.md",
			url + "/README.md",
			url,
		} {
			x, e := asteroidFs.Open(path)
			if e == nil {
				if stat, e := x.Stat(); e == nil && !stat.IsDir() {
					file = x
					servedFilename = path
					break
				}
			}
		}
		if file == nil {
			http.Error(w, "Not Found: "+url, http.StatusNotFound)
			return
		}
		// serve based on file extension
		switch {
		case strings.HasSuffix(servedFilename, ".md"):
			stat, e := file.Stat()
			if e != nil {
				http.Error(w, "can not stat file", http.StatusExpectationFailed)
				return
			}
			content := make([]byte, stat.Size())
			if _, e := file.Read(content); e != nil {
				http.Error(w, "can not read file: error", http.StatusExpectationFailed)
				return
			}
			app.NewTemplatingEngine().
				Set("AsteroidName", asteroidName).
				Set("MainContent", string(content)).
				Set("Config", cfg).
				Render(w, r, "funcs.html", "asteroidGeneric.html")
		case strings.HasSuffix(servedFilename, ".jpg"),
			strings.HasSuffix(servedFilename, ".jpeg"),
			strings.HasSuffix(servedFilename, ".png"),
			strings.HasSuffix(servedFilename, ".gif"),
			strings.HasSuffix(servedFilename, ".svg"):
			http.ServeFileFS(w, r, asteroidFs, servedFilename)
		default:
			http.Error(w, "Unrecognized extension", http.StatusExpectationFailed)
		}
	})
}

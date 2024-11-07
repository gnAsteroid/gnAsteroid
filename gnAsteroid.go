package gnAsteroid

import (
	"embed"
	"html"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gotuna/gotuna"
	"github.com/yalue/merged_fs"
)

const DefaultTheme string = "themes/default.theme"

// Asteroids gravitate around gno.land

// Instead of serving /r/gnoland/blog like gno.land itself,
// gnAsteroid uses gnoweb to serve markdown-based websites.
// The features remain otherwise identical.
//
// Later gnAsteroid can probably also map to
// realm-based wiki to have a local system of files that can
// optionaly be served or backed up/restored to the blockchain.

var (
	asteroidName string // read from cmdLine, or file called .TITLE at root, or "CHANGEME"
	asteroidFs   fs.FS  // used by HandleRootAsMdFile and HandleNotFoundAsFile. This is the main difference with gno.land
)

//go:embed views/*.html
var newViews embed.FS // composed with gnoweb's, using merged_fs

func SetAsteroidFs(asteroid fs.FS) { asteroidFs = asteroid }
func SetAsteroidName(name string)  { asteroidName = name }

// HandleAsteroid can be used to have a serverless gnoweb serving an asteroid as a handler.
// This can be used for example on Vercel.
// @param (asteroid) the fs to serve (normally a tree with some markdown documents)
// @param (theme) if nil, os.DirFS(DefaultTheme) will be used
func HandleAsteroid(asteroid, theme fs.FS, asteroidName_ string, cfg gnoweb.Config) http.Handler {
	SetAsteroidFs(asteroid)
	SetAsteroidName(asteroidName_)
	return MakeApp(slog.Default(), cfg, theme)
}

// MakeApp is separated from HandleAsteroid, mainly because
// cmd/main uses MakeApp(), to reload watched files.
func MakeApp(logger *slog.Logger, cfg gnoweb.Config, themeFs fs.FS) http.Handler {
	gnowebViews, e := fs.Sub(gnoweb.DefaultViewsFiles(), "views")
	if e != nil {
		panic("Could not find gnoweb views: " + e.Error())
	}
	asteroidViews, e := fs.Sub(newViews, "views")
	if e != nil {
		panic("Could not find asteroid views: " + e.Error())
	}
	if themeFs == nil {
		themeFs = os.DirFS(DefaultTheme)
	}
	return gnoweb.MakeAppWithOptions(logger, cfg, gnoweb.Options{
		RootHandler:     HandleRootAsMdFile,
		NotFoundHandler: HandleNotFoundAsFile,
		ThemeFS:         themeFs,
		ViewFS:          merged_fs.NewMergedFS(asteroidViews, gnowebViews),
	}).Router
}

// This RootHandler has gnoweb serve a file called "index.md" or "README.md" when root is requested
// XXX at a glance it seems possible to use the logic from HandleNotFoundAsFile instead.
func HandleRootAsMdFile(logger *slog.Logger, app gotuna.App, cfg *gnoweb.Config) http.Handler {
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
    // Filter out optional Front Matter
    // extracting document Title, if absent Title is the url's path
    // page title is asteroid name, unless defined in Front Matter
    pureMarkdown, _ := ExtractFrontMatter(string(buf))
    pageName := asteroidName 
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Set("AsteroidName", asteroidName).
            Set("AtHome", "1"). // to e.g. disable back_button
            Set("PageName", pageName). 
			Set("Content", string(pureMarkdown)).
			Set("Config", cfg).
			Render(w, r, "asteroid_markdown.html", "funcs.html")
	})
}

// HandleNotFoundAsFile is a fallthrough handler to attempt serving markdown or images
func HandleNotFoundAsFile(logger *slog.Logger, app gotuna.App, cfg *gnoweb.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		url = html.UnescapeString(url)
		url = strings.TrimPrefix(url, "/") // with embedfs, need to remove the leading /
		url = filepath.Clean(url)

		if strings.Contains(url, "..") {
			app.NewTemplatingEngine().
				Set("AsteroidName", asteroidName).
				Set("Config", cfg).
				Render(w, r, "asteroid_403.html", "funcs.html")
            return
		}
		var file fs.File // when nil, means still not found
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
            // Filter out optional Front Matter
            // extracting document Title, if absent Title is the url's path
            pureMarkdown, kv := ExtractFrontMatter(string(content))
            pageName := strings.TrimSuffix(url, ".md") // e.g. "subdir/deep/blue" (without .md)
            if title, has := kv["title"]; has {
                pageName = title
            }
			app.NewTemplatingEngine().
				Set("AsteroidName", asteroidName).
                Set("AtHome", "0"). // to e.g. allow back_button
                Set("PageName", pageName). 
				Set("Content", string(pureMarkdown)).
				Set("Config", cfg).
				Render(w, r, "funcs.html", "asteroid_markdown.html")
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

// given a `content` supposedly in markdown,
// return `pureMarkdown` which is the same as `content` without a Front Matter
// header.
//
// Key-values will be stored to a map[string]string, the keys are lower-cased.
//
// Here are two examples of Front Matter:
//
// ---
// title: Document Title
// date: YYYY-MM-DD
// description: A brief description.
// tags: [tag1, tag2]
// author: Your Name
// ---
// And the rest is markdown.
//
//
// ---title: this is a one line front matter---
// And the rest is markdown.
//
//
func ExtractFrontMatter(content string) (pureMarkdown string, kv map[string]string) {
	kv = make(map[string]string)
    reOneLine := regexp.MustCompile(`^---([^:\s]+)\s*:\s*(.*)---\s*$`)
    if m := reOneLine.FindStringSubmatch(content[:strings.IndexByte(content, '\n')]); m != nil {
        kv[strings.ToLower(m[1])] = m[2]
        pureMarkdown = content[strings.IndexByte(content, '\n')+1:]
    } else if !strings.HasPrefix(content, "---\n") {
        // Front Matter
        pureMarkdown = content
    } else {
        inFrontMatter := true
        reKv := regexp.MustCompile(`^(\w+):\s*(.*)$`)
        for _, line := range strings.Split(content, "\n")[1:] {
            if inFrontMatter {
                if line == "---" {
                    inFrontMatter = false
                } else if m := reKv.FindStringSubmatch(line); m != nil {
                    kv[strings.ToLower(m[1])] = m[2]
                }
            } else {
                pureMarkdown += line + "\n"
            }
        }
    }
    return 
}


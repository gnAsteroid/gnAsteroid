package gnAsteroid

// gnoweb.go and gnoweb_test are gnoweb 1.0.
//
// They are september 2024 version of gnolang/gno/gno.land/pkg/gnoweb/gnoweb.go
// Initial plan was to depend on gnoweb 2.0, but after digging a little bit,
// it would take ten times the time (all components, views would need to be
// rewritten. Besides, gnoweb 1.0 was a fine piece of software, if gno can work
// with both gnoweb 2.0 and gnAsteroid it may be good for gno).

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gorilla/mux"
	"github.com/gotuna/gotuna"

	"github.com/gnAsteroid/gnAsteroid/static"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
)

const (
	qFileStr = "vm/qfile"
)

//go:embed views/*
var defaultViewsFiles embed.FS // getter: DefaultViewsFiles()

type Config struct {
	RemoteAddr    string
	ViewsDir      string
	HelpChainID   string
	HelpRemote    string
	WithAnalytics bool
}

type Options struct {
	Aliases         map[string]string
	Redirects       map[string]string
	RootHandler     func(*slog.Logger, gotuna.App, *Config) http.Handler // if set, is called after aliases and redirects (if either defines "/", this handler thus will never be called)
	NotFoundHandler func(*slog.Logger, gotuna.App, *Config) http.Handler // if set, will be used instead of default notFoundHandler
	ViewFS          fs.FS                                                // if set, has precedence over ViewsDir
	ThemeFS         fs.FS                                                // optional theme directory (containing css/, img/ and font/)
}

func NewDefaultConfig() *Config {
	return &Config{
		RemoteAddr:    "127.0.0.1:26657",
		ViewsDir:      "",
		HelpChainID:   "dev",
		HelpRemote:    "127.0.0.1:26657",
		WithAnalytics: false,
	}
}

func MakeGnowebApp(logger *slog.Logger, cfg *Config) gotuna.App {
	return MakeGnowebAppWithOptions(logger, cfg, Options{})
}

func MakeGnowebAppWithOptions(logger *slog.Logger, cfg *Config, opts Options) gotuna.App {
	var viewFiles fs.FS
	var themeFiles fs.FS = opts.ThemeFS

	if opts.ViewFS != nil {
		viewFiles = opts.ViewFS
	} else if cfg.ViewsDir != "" {
		viewFiles = os.DirFS(cfg.ViewsDir)
	} else {
		var err error
		viewFiles, err = fs.Sub(defaultViewsFiles, "views")
		if err != nil {
			panic("unable to get views directory from embed fs: " + err.Error())
		}
	}

	app := gotuna.App{
		ViewFiles: viewFiles,
		Router:    gotuna.NewMuxRouter(),
		Static:    static.EmbeddedStatic,
	}

	for from, to := range opts.Aliases {
		app.Router.Handle(from, handlerRealmAlias(logger, app, cfg, to))
	}
	for from, to := range opts.Redirects {
		app.Router.Handle(from, handlerRedirect(logger, app, cfg, to))
	}
	if opts.RootHandler != nil {
		app.Router.Handle("/", opts.RootHandler(logger, app, cfg))
	}
	// realm routes
	// NOTE: see rePathPart.
	app.Router.Handle("/r/{rlmname:[a-z][a-z0-9_]*(?:/[a-z][a-z0-9_]*)+}/{filename:(?:(?:.*\\.(?:gno|md|txt|mod)$)|(?:LICENSE$))?}", handlerRealmFile(logger, app, cfg))
	app.Router.Handle("/r/{rlmname:[a-z][a-z0-9_]*(?:/[a-z][a-z0-9_]*)+}", handlerRealmMain(logger, app, cfg))
	app.Router.Handle("/r/{rlmname:[a-z][a-z0-9_]*(?:/[a-z][a-z0-9_]*)+}:{querystr:.*}", handlerRealmRender(logger, app, cfg))
	app.Router.Handle("/p/{filepath:.*}", handlerPackageFile(logger, app, cfg))

	// other
	app.Router.Handle("/faucet", handlerFaucet(logger, app, cfg))

	if themeFiles != nil {
		// alternative styling for css, font and img.
		// if assets are not found here, static/* assets are
		// still handled by the next line after this if-block
		// which works as a fallthrough.
		app.Router.Handle("/static/{path:(?:css|font|img)/.+}", handlerStaticFile(logger, app, cfg, themeFiles))
	}
	app.Router.Handle("/static/{path:.+}", handlerStaticFile(logger, app, cfg, app.Static))
	app.Router.Handle("/favicon.ico", handlerFavicon(logger, app, cfg))

	// api
	app.Router.Handle("/status.json", handlerStatusJSON(logger, app, cfg))

	if opts.NotFoundHandler != nil {
		app.Router.NotFoundHandler = opts.NotFoundHandler(logger, app, cfg)
	} else {
		app.Router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.RequestURI
			handleNotFound(logger, app, cfg, path, w, r)
		})
	}

	return app
}

// handlerRealmAlias is used to render official pages from realms.
// url is intended to be shorter.
// UX is intended to be more minimalistic.
// A link to the realm realm is added.
func handlerRealmAlias(logger *slog.Logger, app gotuna.App, cfg *Config, rlmpath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rlmfullpath := "gno.land" + rlmpath
		querystr := "" // XXX: "?gnoweb-alias=1"
		parts := strings.Split(rlmpath, ":")
		switch len(parts) {
		case 1: // continue
		case 2: // r/realm:querystr
			rlmfullpath = "gno.land" + parts[0]
			querystr = parts[1] + querystr
		default:
			panic("should not happen")
		}
		rlmname := strings.TrimPrefix(rlmfullpath, "gno.land/r/")
		qpath := "vm/qrender"
		data := []byte(fmt.Sprintf("%s:%s", rlmfullpath, querystr))
		res, err := makeRequest(logger, cfg, qpath, data)
		if err != nil {
			writeError(logger, w, fmt.Errorf("gnoweb failed to query gnoland: %w", err))
			return
		}

		queryParts := strings.Split(querystr, "/")
		pathLinks := []pathLink{}
		for i, part := range queryParts {
			pathLinks = append(pathLinks, pathLink{
				URL:  "/r/" + rlmname + ":" + strings.Join(queryParts[:i+1], "/"),
				Text: part,
			})
		}

		tmpl := app.NewTemplatingEngine()
		// XXX: extract title from realm's output
		// XXX: extract description from realm's output
		tmpl.Set("RealmName", rlmname)
		tmpl.Set("RealmPath", rlmpath)
		tmpl.Set("Query", querystr)
		tmpl.Set("PathLinks", pathLinks)
		tmpl.Set("Contents", string(res.Data))
		tmpl.Set("Config", cfg)
		tmpl.Set("IsAlias", true)
		tmpl.Render(w, r, "realm_render.html", "funcs.html")
	})
}

func handlerFaucet(logger *slog.Logger, app gotuna.App, cfg *Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Set("Config", cfg).
			Render(w, r, "faucet.html", "funcs.html")
	})
}

func handlerStatusJSON(logger *slog.Logger, app gotuna.App, cfg *Config) http.Handler {
	startedAt := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ret struct {
			Gnoland struct {
				Connected bool    `json:"connected"`
				Error     *string `json:"error,omitempty"`
				Height    *int64  `json:"height,omitempty"`
				// processed txs
				// active connections

				Version *string `json:"version,omitempty"`
				// Uptime    *float64 `json:"uptime-seconds,omitempty"`
				// Goarch    *string  `json:"goarch,omitempty"`
				// Goos      *string  `json:"goos,omitempty"`
				// GoVersion *string  `json:"go-version,omitempty"`
				// NumCPU    *int     `json:"num_cpu,omitempty"`
			} `json:"gnoland"`
			Website struct {
				// Version string  `json:"version"`
				Uptime    float64 `json:"uptime-seconds"`
				Goarch    string  `json:"goarch"`
				Goos      string  `json:"goos"`
				GoVersion string  `json:"go-version"`
				NumCPU    int     `json:"num_cpu"`
			} `json:"website"`
		}
		ret.Website.Uptime = time.Since(startedAt).Seconds()
		ret.Website.Goarch = runtime.GOARCH
		ret.Website.Goos = runtime.GOOS
		ret.Website.NumCPU = runtime.NumCPU()
		ret.Website.GoVersion = runtime.Version()

		ret.Gnoland.Connected = true
		res, err := makeRequest(logger, cfg, ".app/version", []byte{})
		if err != nil {
			ret.Gnoland.Connected = false
			errmsg := err.Error()
			ret.Gnoland.Error = &errmsg
		} else {
			version := string(res.Value)
			ret.Gnoland.Version = &version
			ret.Gnoland.Height = &res.Height
		}

		out, _ := json.MarshalIndent(ret, "", "  ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	})
}

func handlerRedirect(logger *slog.Logger, app gotuna.App, cfg *Config, to string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, http.StatusFound)
		tmpl := app.NewTemplatingEngine()
		tmpl.Set("To", to)
		tmpl.Set("Config", cfg)
		tmpl.Render(w, r, "redirect.html", "funcs.html")
	})
}

func handlerRealmMain(logger *slog.Logger, app gotuna.App, cfg *Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		rlmname := vars["rlmname"]
		rlmpath := "gno.land/r/" + rlmname
		query := r.URL.Query()

		logger.Info("handling", "name", rlmname, "path", rlmpath)
		if query.Has("help") {
			// Render function helper.
			funcName := query.Get("__func")
			qpath := "vm/qfuncs"
			data := []byte(rlmpath)
			res, err := makeRequest(logger, cfg, qpath, data)
			if err != nil {
				writeError(logger, w, fmt.Errorf("request failed: %w", err))
				return
			}
			var fsigs vm.FunctionSignatures
			amino.MustUnmarshalJSON(res.Data, &fsigs)
			// Fill fsigs with query parameters.
			for i := range fsigs {
				fsig := &(fsigs[i])
				for j := range fsig.Params {
					param := &(fsig.Params[j])
					value := query.Get(param.Name)
					param.Value = value
				}
			}
			// Render template.
			tmpl := app.NewTemplatingEngine()
			tmpl.Set("FuncName", funcName)
			tmpl.Set("RealmPath", rlmpath)
			tmpl.Set("DirPath", pathOf(rlmpath))
			tmpl.Set("FunctionSignatures", fsigs)
			tmpl.Set("Config", cfg)
			tmpl.Render(w, r, "funcs.html", "realm_help.html")
		} else {
			// Ensure realm exists. TODO optimize.
			qpath := qFileStr
			data := []byte(rlmpath)
			_, err := makeRequest(logger, cfg, qpath, data)
			if err != nil {
				writeError(logger, w, errors.New("error querying realm package, remote="+cfg.RemoteAddr))
				return
			}
			// Render blank query path, /r/REALM:.
			handleRealmRender(logger, app, cfg, w, r)
		}
	})
}

type pathLink struct {
	URL  string
	Text string
}

func handlerRealmRender(logger *slog.Logger, app gotuna.App, cfg *Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRealmRender(logger, app, cfg, w, r)
	})
}

func handleRealmRender(logger *slog.Logger, app gotuna.App, cfg *Config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rlmname := vars["rlmname"]
	rlmpath := "gno.land/r/" + rlmname
	querystr := vars["querystr"]
	if r.URL.Path == "/r/"+rlmname+":" {
		// Redirect to /r/REALM if querypath is empty.
		http.Redirect(w, r, "/r/"+rlmname, http.StatusFound)
		return
	}
	qpath := "vm/qrender"
	data := []byte(fmt.Sprintf("%s:%s", rlmpath, querystr))
	res, err := makeRequest(logger, cfg, qpath, data)
	if err != nil {
		// XXX hack
		if strings.Contains(err.Error(), "Render not declared") {
			res = &abci.ResponseQuery{}
			res.Data = []byte("realm package has no Render() function")
		} else {
			writeError(logger, w, err)
			return
		}
	}

	dirdata := []byte(rlmpath)
	dirres, err := makeRequest(logger, cfg, qFileStr, dirdata)
	if err != nil {
		writeError(logger, w, err)
		return
	}
	hasReadme := bytes.Contains(append(dirres.Data, '\n'), []byte("README.md\n"))

	// linkify querystr.
	queryParts := strings.Split(querystr, "/")
	pathLinks := []pathLink{}
	for i, part := range queryParts {
		pathLinks = append(pathLinks, pathLink{
			URL:  "/r/" + rlmname + ":" + strings.Join(queryParts[:i+1], "/"),
			Text: part,
		})
	}
	// Render template.
	tmpl := app.NewTemplatingEngine()
	// XXX: extract title from realm's output
	// XXX: extract description from realm's output
	tmpl.Set("RealmName", rlmname)
	tmpl.Set("RealmPath", rlmpath)
	tmpl.Set("Query", querystr)
	tmpl.Set("PathLinks", pathLinks)
	tmpl.Set("Contents", string(res.Data))
	tmpl.Set("Config", cfg)
	tmpl.Set("HasReadme", hasReadme)
	tmpl.Render(w, r, "realm_render.html", "funcs.html")
}

func handlerRealmFile(logger *slog.Logger, app gotuna.App, cfg *Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		diruri := "gno.land/r/" + vars["rlmname"]
		filename := vars["filename"]
		renderPackageFile(logger, app, cfg, w, r, diruri, filename)
	})
}

func handlerPackageFile(logger *slog.Logger, app gotuna.App, cfg *Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pkgpath := "gno.land/p/" + vars["filepath"]
		diruri, filename := SplitFilepath(pkgpath)
		if filename == "" && diruri == pkgpath {
			// redirect to diruri + "/"
			http.Redirect(w, r, "/p/"+vars["filepath"]+"/", http.StatusFound)
			return
		}
		renderPackageFile(logger, app, cfg, w, r, diruri, filename)
	})
}

func renderPackageFile(logger *slog.Logger, app gotuna.App, cfg *Config, w http.ResponseWriter, r *http.Request, diruri string, filename string) {
	if filename == "" {
		// Request is for a folder.
		qpath := qFileStr
		data := []byte(diruri)
		res, err := makeRequest(logger, cfg, qpath, data)
		if err != nil {
			writeError(logger, w, err)
			return
		}
		files := strings.Split(string(res.Data), "\n")
		// Render template.
		tmpl := app.NewTemplatingEngine()
		tmpl.Set("DirURI", diruri)
		tmpl.Set("DirPath", pathOf(diruri))
		tmpl.Set("Files", files)
		tmpl.Set("Config", cfg)
		tmpl.Render(w, r, "package_dir.html", "funcs.html")
	} else {
		// Request is for a file.
		filepath := diruri + "/" + filename
		qpath := qFileStr
		data := []byte(filepath)
		res, err := makeRequest(logger, cfg, qpath, data)
		if err != nil {
			writeError(logger, w, err)
			return
		}
		// Render template.
		tmpl := app.NewTemplatingEngine()
		tmpl.Set("DirURI", diruri)
		tmpl.Set("DirPath", pathOf(diruri))
		tmpl.Set("FileName", filename)
		tmpl.Set("FileContents", string(res.Data))
		tmpl.Set("Config", cfg)
		tmpl.Render(w, r, "package_file.html", "funcs.html")
	}
}

func makeRequest(log *slog.Logger, cfg *Config, qpath string, data []byte) (res *abci.ResponseQuery, err error) {
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	remote := cfg.RemoteAddr
	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP client, %w", err)
	}

	qres, err := cli.ABCIQueryWithOptions(
		qpath, data, opts2)
	if err != nil {
		log.Error("request error", "path", qpath, "error", err)
		return nil, fmt.Errorf("unable to query path %q: %w", qpath, err)
	}
	if qres.Response.Error != nil {
		log.Error("response error", "path", qpath, "log", qres.Response.Log)
		return nil, qres.Response.Error
	}
	return &qres.Response, nil
}

func handlerStaticFile(logger *slog.Logger, app gotuna.App, cfg *Config, filesystem fs.FS) http.Handler {
	fs := http.FS(filesystem)
	fileapp := http.StripPrefix("/static", http.FileServer(fs))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fpath := filepath.Clean(vars["path"])
		logger.Debug("Serving ", "static", fpath)
		f, err := fs.Open(fpath)
		if os.IsNotExist(err) || err != nil || f == nil {
			logger.Debug("Not found: " + fpath)
			handleNotFound(logger, app, cfg, fpath, w, r)
			return
		}
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			handleNotFound(logger, app, cfg, fpath, w, r)
			return
		}

		// TODO: ModTime doesn't work for embed?
		// w.Header().Set("ETag", fmt.Sprintf("%x", stat.ModTime().UnixNano()))
		// w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%s", "31536000"))
		// Discussion about support of ETag in embedfs: https://github.com/golang/go/issues/60940
		fileapp.ServeHTTP(w, r)
	})
}

func handlerFavicon(logger *slog.Logger, app gotuna.App, cfg *Config) http.Handler {
	fs := http.FS(app.Static)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fpath := "img/favicon.ico"
		f, err := fs.Open(fpath)
		if os.IsNotExist(err) {
			handleNotFound(logger, app, cfg, fpath, w, r)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=604800") // 7d
		io.Copy(w, f)
	})
}

func handleNotFound(logger *slog.Logger, app gotuna.App, cfg *Config, path string, w http.ResponseWriter, r *http.Request) {
	// decode path for non-ascii characters
	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		logger.Error("failed to decode path", err)
		decodedPath = path
	}
	w.WriteHeader(http.StatusNotFound)
	app.NewTemplatingEngine().
		Set("title", "Not found").
		Set("path", decodedPath).
		Set("Config", cfg).
		Render(w, r, "404.html", "funcs.html")
}

func writeError(logger *slog.Logger, w http.ResponseWriter, err error) {
	if details := errors.Unwrap(err); details != nil {
		logger.Error("handler", "error", err, "details", details)
	} else {
		logger.Error("handler", "error:", err)
	}

	// XXX: writeError should return an error page template.
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

func pathOf(diruri string) string {
	parts := strings.Split(diruri, "/")
	if parts[0] == "gno.land" {
		return "/" + strings.Join(parts[1:], "/")
	}

	panic(fmt.Sprintf("invalid dir-URI %q", diruri))
}

// Getter to defaultViewsFiles
func DefaultViewsFiles() embed.FS {
	return defaultViewsFiles
}

// Splits a path into the dirpath and filename.
// previously in tm2/pkg/memfile.go
func SplitFilepath(filepath string) (dirpath string, filename string) {
	licenseName := "LICENSE"
	parts := strings.Split(filepath, "/")
	if len(parts) == 1 {
		return parts[0], ""
	}

	switch last := parts[len(parts)-1]; {
	case strings.Contains(last, "."):
		return strings.Join(parts[:len(parts)-1], "/"), last
	case last == "":
		return strings.Join(parts[:len(parts)-1], "/"), ""
	case last == licenseName:
		return strings.Join(parts[:len(parts)-1], "/"), licenseName
	}

	return strings.Join(parts, "/"), ""
}

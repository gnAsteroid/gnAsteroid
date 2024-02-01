package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	osm "github.com/gnolang/gno/pkgs/os"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gorilla/mux"
	"github.com/gotuna/gotuna"

	"github.com/gnolang/gno/pkgs/sdk/vm"       // for error types
	"github.com/grepsuzette/gnAsteroid/static" // for static files
	// "github.com/gnolang/gno/pkgs/sdk"       // for baseapp (info, status)
)

const (
	qFileStr = "vm/qfile"
)

var flags struct {
	asteroidDir  string
	asteroidName string
	bindAddr     string
	remoteAddr   string
	styleDir     string
	helpChainID  string
	helpRemote   string
}

var startedAt time.Time

func init() {
	flag.StringVar(&flags.asteroidDir, "asteroid-dir", "", "wiki directory location. [Mandatory!]")
	flag.StringVar(&flags.asteroidName, "asteroid-name", "CHANGEME", "the asteroid name (website title). Default=CHANGEME")
	flag.StringVar(&flags.bindAddr, "bind", "0.0.0.0:8888", "server listening address")
	flag.StringVar(&flags.styleDir, "style-dir", "", "style directory (css, js, img). Default=<internal>")
	flag.StringVar(&flags.remoteAddr, "remote", "127.0.0.1:26657", "remote gnoland node address")
	flag.StringVar(&flags.helpChainID, "help-chainid", "dev", "help page's chainid")
	flag.StringVar(&flags.helpRemote, "help-remote", "127.0.0.1:26657", "help page's remote addr")
	startedAt = time.Now()
}

func makeApp() gotuna.App {
	app := gotuna.App{
		Static:    static.EmbeddedStatic,
		ViewFiles: static.EmbeddedStatic, // prefix is "views/"
		Router:    gotuna.NewMuxRouter(),
	}
	app.Router.Handle("/", handlerHome(app))
	app.Router.Handle("/r/demo/boards:gnolang/6", handlerRedirect(app))
	// NOTE: see rePathPart.
	app.Router.Handle("/r/{rlmname:[a-z][a-z0-9_]*(?:/[a-z][a-z0-9_]*)+}/{filename:(?:.*\\.(?:gno|md|txt)$)?}", handlerRealmFile(app))
	app.Router.Handle("/r/{rlmname:[a-z][a-z0-9_]*(?:/[a-z][a-z0-9_]*)+}", handlerRealmMain(app))
	app.Router.Handle("/r/{rlmname:[a-z][a-z0-9_]*(?:/[a-z][a-z0-9_]*)+}:{querystr:.*}", handlerRealmRender(app))
	app.Router.Handle("/p/{filepath:.*}", handlerPackageFile(app))
	if osm.DirExists(flags.styleDir) {
		app.Router.Handle("/static/{path:(?:css|font|img)/.+}", handlerAsset(app))
	}
	// else the style is simply in embed.FS as well:
	app.Router.Handle("/static/{path:.+}", handlerStaticFile(app))
	app.Router.Handle("/favicon.ico", handlerFavicon(app))
	app.Router.Handle("/status.json", handlerStatusJSON(app))
	app.Router.Handle("/{anything:.*(?:\\.md)?}", handlerAnything(app))
	return app
}

func main() {
	flag.Parse()
	fmt.Printf("Serving %s on http://%s\n", flags.asteroidDir, flags.bindAddr)
	server := &http.Server{
		Addr:              flags.bindAddr,
		ReadHeaderTimeout: 60 * time.Second,
		Handler:           makeApp().Router,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server stopped with error: %+v\n", err)
	}
}

func handlerHome(app gotuna.App) http.Handler {
	md := filepath.Join(flags.asteroidDir, "index.md")
	homeContent := osm.MustReadFile(md)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Set("AsteroidName", flags.asteroidName).
			Set("HomeContent", string(homeContent)).
			Render(w, r, "views/home.html", "views/funcs.html")
	})
}

func handlerAnything(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := filepath.Join(flags.asteroidDir, mux.Vars(r)["anything"])
		if osm.DirExists(file) && osm.FileExists(file+"/index.md") {
			// it's a dir
			file = file + "/index.md"
		} else if osm.FileExists(file + ".md") {
			file = file + ".md"
		} else if osm.FileExists(file) { // exists and file or dir
		} else if osm.DirExists(file) { // exists and file or dir
			http.Error(w, "Teapot: missing index.md", http.StatusTeapot)
			return
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		if strings.Contains(file, "..") {
			app.NewTemplatingEngine().
				Set("AsteroidName", flags.asteroidName).
				Render(w, r, "views/403.html", "views/funcs.html")
		}
		// serve the file, based of its extension
		switch {
		case strings.HasSuffix(file, ".md"):
			content := osm.MustReadFile(file)
			app.NewTemplatingEngine().
				Set("AsteroidName", flags.asteroidName).
				Set("MainContent", string(content)).
				Render(w, r, "views/generic.html", "views/funcs.html")
		case strings.HasSuffix(file, ".jpg") || strings.HasSuffix(file, ".jpeg"):
			http.ServeFile(w, r, file)
		default:
			if osm.FileExists(file) {
				w.Write([]byte("Unrecognized extension"))
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		}
	})
}

func handlerStatusJSON(app gotuna.App) http.Handler {
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
		res, err := makeRequest(".app/version", []byte{})
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

// XXX temporary.
func handlerRedirect(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/r/boards:gnolang/3", http.StatusFound)
		app.NewTemplatingEngine().
			Set("AsteroidName", flags.asteroidName).
			Render(w, r, "views/home.html", "views/funcs.html")
	})
}

func handlerRealmMain(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		rlmname := vars["rlmname"]
		rlmpath := "gno.land/r/" + rlmname
		query := r.URL.Query()
		if query.Has("help") {
			// Render function helper.
			funcName := query.Get("__func")
			qpath := "vm/qfuncs"
			data := []byte(rlmpath)
			res, err := makeRequest(qpath, data)
			if err != nil {
				writeError(w, err)
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
			tmpl.Set("AsteroidName", flags.asteroidName)
			tmpl.Set("FuncName", funcName)
			tmpl.Set("RealmPath", rlmpath)
			// tmpl.Set("FormBodyType", query.Get("body.type")) // when "textarea", <textarea> will be used instead of <input>
			tmpl.Set("FormBodyType", "textarea") // when "textarea", <textarea> will be used instead of <input>
			tmpl.Set("Remote", flags.helpRemote)
			tmpl.Set("ChainID", flags.helpChainID)
			tmpl.Set("DirPath", pathOf(rlmpath))
			tmpl.Set("FunctionSignatures", fsigs)
			tmpl.Render(w, r, "views/realm_help.html", "views/funcs.html")
		} else {
			// Ensure realm exists. TODO optimize.
			qpath := qFileStr
			data := []byte(rlmpath)
			_, err := makeRequest(qpath, data)
			if err != nil {
				writeError(w, errors.New("error querying realm package"))
				return
			}
			// Render blank query path, /r/REALM:.
			handleRealmRender(app, w, r)
		}
	})
}

type pathLink struct {
	URL  string
	Text string
}

func handlerRealmRender(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRealmRender(app, w, r)
	})
}

func handleRealmRender(app gotuna.App, w http.ResponseWriter, r *http.Request) {
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
	data := []byte(fmt.Sprintf("%s\n%s", rlmpath, querystr))
	res, err := makeRequest(qpath, data)
	if err != nil {
		// XXX hack
		if strings.Contains(err.Error(), "Render not declared") {
			res = &abci.ResponseQuery{}
			res.Data = []byte("realm package `" + rlmname + "` has no Render() function")
		} else {
			writeError(w, err)
			return
		}
	}
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
	tmpl.Set("AsteroidName", flags.asteroidName)
	tmpl.Set("RealmName", rlmname)
	tmpl.Set("RealmPath", rlmpath)
	tmpl.Set("Query", querystr)
	tmpl.Set("PathLinks", pathLinks)
	tmpl.Set("Contents", string(res.Data))
	tmpl.Render(w, r, "views/realm_render.html", "views/funcs.html")
}

func handlerRealmFile(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		diruri := "gno.land/r/" + vars["rlmname"]
		filename := vars["filename"]
		renderPackageFile(app, w, r, diruri, filename)
	})
}

func handlerPackageFile(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pkgpath := "gno.land/p/" + vars["filepath"]
		diruri, filename := std.SplitFilepath(pkgpath)
		if filename == "" && diruri == pkgpath {
			// redirect to diruri + "/"
			http.Redirect(w, r, "/p/"+vars["filepath"]+"/", http.StatusFound)
			return
		}
		renderPackageFile(app, w, r, diruri, filename)
	})
}

func renderPackageFile(app gotuna.App, w http.ResponseWriter, r *http.Request, diruri string, filename string) {
	if filename == "" {
		// Request is for a folder.
		qpath := qFileStr
		data := []byte(diruri)
		res, err := makeRequest(qpath, data)
		if err != nil {
			writeError(w, err)
			return
		}
		files := strings.Split(string(res.Data), "\n")
		// Render template.
		tmpl := app.NewTemplatingEngine()
		tmpl.Set("AsteroidName", flags.asteroidName)
		tmpl.Set("DirURI", diruri)
		tmpl.Set("DirPath", pathOf(diruri))
		tmpl.Set("Files", files)
		tmpl.Render(w, r, "views/package_dir.html", "views/funcs.html")
	} else {
		// Request is for a file.
		filepath := diruri + "/" + filename
		qpath := qFileStr
		data := []byte(filepath)
		res, err := makeRequest(qpath, data)
		if err != nil {
			writeError(w, err)
			return
		}
		// Render template.
		tmpl := app.NewTemplatingEngine()
		tmpl.Set("AsteroidName", flags.asteroidName)
		tmpl.Set("DirURI", diruri)
		tmpl.Set("DirPath", pathOf(diruri))
		tmpl.Set("FileName", filename)
		tmpl.Set("FileContents", string(res.Data))
		tmpl.Render(w, r, "views/package_file.html", "views/funcs.html")
	}
}

func makeRequest(qpath string, data []byte) (res *abci.ResponseQuery, err error) {
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	remote := flags.remoteAddr
	cli := client.NewHTTP(remote, "/websocket")
	qres, err := cli.ABCIQueryWithOptions(
		qpath, data, opts2)
	if err != nil {
		return nil, err
	}
	if qres.Response.Error != nil {
		fmt.Printf("Log: %s\n",
			qres.Response.Log)
		return nil, qres.Response.Error
	}
	return &qres.Response, nil
}

func handlerAsset(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := flags.styleDir
		file := filepath.Join(s, mux.Vars(r)["path"])
		if !osm.FileExists(file) || osm.DirExists(file) {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			// serve the file, based of its extension
			http.ServeFile(w, r, file)
		}
	})
}

func handlerStaticFile(app gotuna.App) http.Handler {
	fs := http.FS(app.Static)
	fileapp := http.StripPrefix("/static", http.FileServer(fs))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fpath := filepath.Clean(vars["path"])
		f, err := fs.Open(fpath)
		if os.IsNotExist(err) {
			handleNotFound(app, fpath, w, r)
			return
		}
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			handleNotFound(app, fpath, w, r)
			return
		}

		// TODO: ModTime doesn't work for embed?
		// w.Header().Set("ETag", fmt.Sprintf("%x", stat.ModTime().UnixNano()))
		// w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%s", "31536000"))
		fileapp.ServeHTTP(w, r)
	})
}

func handlerFavicon(app gotuna.App) http.Handler {
	if osm.DirExists(flags.styleDir) && osm.FileExists(flags.styleDir+"/img/favicon.ico") {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, flags.styleDir+"/img/favicon.ico")
		})
	} else {
		fs := http.FS(app.Static)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fpath := "img/favicon.ico"
			f, err := fs.Open(fpath)
			if os.IsNotExist(err) {
				handleNotFound(app, fpath, w, r)
				return
			}
			w.Header().Set("Content-Type", "image/x-icon")
			w.Header().Set("Cache-Control", "public, max-age=604800") // 7d
			io.Copy(w, f)
		})
	}
}

func handleNotFound(app gotuna.App, path string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	app.NewTemplatingEngine().
		Set("AsteroidName", flags.asteroidName).
		Set("title", "Not found").
		Set("path", path).
		Render(w, r, "views/404.html", "views/funcs.html")
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

func pathOf(diruri string) string {
	parts := strings.Split(diruri, "/")
	if parts[0] == "gno.land" {
		return "/" + strings.Join(parts[1:], "/")
	} else {
		panic(fmt.Sprintf("invalid dir-URI %q", diruri))
	}
}

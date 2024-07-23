package gnAsteroid

import (
	"embed"
	"flag"
	"fmt"
	"html"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dietsche/rfsnotify"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gotuna/gotuna"
	"github.com/yalue/merged_fs"
	"go.uber.org/zap/zapcore"
	"gopkg.in/fsnotify.v1"
)

var (
	asteroidName string // read from cmdLine, or file called .TITLE at root, or "CHANGEME"
	asteroidDir  string // asteroidDir will be read and become asteroidFs
	asteroidFs   fs.FS  // ultimately used by HandlerHome and HandlerAnything
	styleDir     string
	bindAddr     string // not used when HandleAsteroid()
)

//go:embed views/*.html
var newViews embed.FS // composed with gnoweb's, using merged_fs

//go:embed default-style/*
var defaultEmbedStyle embed.FS

func SetAsteroidFs(asteroid fs.FS) { asteroidFs = asteroid }
func SetAsteroidName(name string)  { asteroidName = name }
func DefaultStyle() fs.FS {
	if x, e := fs.Sub(defaultEmbedStyle, "default-style"); e == nil {
		return x
	} else {
		panic("bad go:embed? " + e.Error())
	}
}

// encapsulate a serverless gnoweb serving an asteroid as an http.Handler
// It can then be served, for example on Vercel. nil style uses
// defaultEmbedStyle.
func HandleAsteroid(asteroid, style fs.FS, asteroidName_ string, cfg gnoweb.Config) http.Handler {
	SetAsteroidFs(asteroid)
	SetAsteroidName(asteroidName_)
	return MakeApp(slog.Default(), cfg, style)
}

func MakeApp(logger *slog.Logger, cfg gnoweb.Config, styleFs fs.FS) http.Handler {
	gnowebViews, e := fs.Sub(gnoweb.DefaultViewsFiles(), "views")
	if e != nil {
		panic("Could not find gnoweb views: " + e.Error())
	}
	if styleFs == nil {
		styleFs = defaultEmbedStyle
	}
	return gnoweb.MakeAppWithOptions(logger, cfg, gnoweb.Options{
		RootHandler:     HandlerRoot,
		NotFoundHandler: HandlerAnything,
		StyleFS:         styleFs,
		ViewFS:          merged_fs.NewMergedFS(gnowebViews, newViews),
	}).Router
}

// main() for cli gnAsteroid.
func main() {
	zapLogger := log.NewZapConsoleLogger(os.Stdout, zapcore.DebugLevel)
	logger := log.ZapLoggerToSlog(zapLogger)

	cfg, e := parseArgs(os.Args[1:], logger)
	if e != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", e)
		os.Exit(1)
	}
	logger.Info(fmt.Sprintf("Serving %s on http://%s", asteroidDir, bindAddr))
	styleFs := StyleFsFrom(styleDir)
	server := &http.Server{
		Addr:              bindAddr,
		ReadHeaderTimeout: 60 * time.Second,
		Handler:           MakeApp(logger, cfg, styleFs),
	}

	// SIGUSR1 -> invalidate all templates
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)
	go func() {
		for {
			sig := <-sigChan
			switch sig {
			case syscall.SIGUSR1:
				fmt.Println("received SIGUSR1.")
				server.Handler = MakeApp(logger, cfg, styleFs)
			}
		}
	}()

	// Watch over asteroidDir for any change -> invalidate all templates
	if watcher, err := rfsnotify.NewWatcher(); err == nil {
		defer watcher.Close()
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						logger.Info("Reloading, modified: " + event.Name)
						server.Handler = MakeApp(logger, cfg, styleFs)
					}
				}
			}
		}()
		watcher.AddRecursive(asteroidDir)
		if styleDir != "" {
			watcher.AddRecursive(styleDir)
		}
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error(fmt.Sprintf("HTTP server stopped with error: %+v\n", err))
	}
	zapLogger.Sync()
}

func parseArgs(args []string, logger *slog.Logger) (gnoweb.Config, error) {
	cfg := gnoweb.NewDefaultConfig()
	fs := flag.NewFlagSet("gnoweb", flag.ContinueOnError)
	// gnAsteroid flags
	fs.StringVar(&asteroidDir, "asteroid-dir", "", "wiki directory location. [Mandatory!]")
	fs.StringVar(&asteroidName, "asteroid-name", "CHANGEME", "the asteroid name (website title). read from .TITLE, or CHANGEME")
	fs.StringVar(&styleDir, "style-dir", "", "style directory (css, js, img). Default is '', meaning \"use embed 'default-style/'\"")
	fs.StringVar(&bindAddr, "bind", "0.0.0.0:8888", "server listening address")
	// gnoweb flags
	fs.StringVar(&cfg.RemoteAddr, "remote", "127.0.0.1:26657", "remote gnoland node address")
	fs.StringVar(&cfg.CaptchaSite, "captcha-site", "", "recaptcha site key (if empty, captcha are disabled)")
	fs.StringVar(&cfg.FaucetURL, "faucet-url", "http://localhost:5050", "faucet server URL")
	fs.StringVar(&cfg.HelpChainID, "help-chainid", "dev", "help page's chainid")
	fs.StringVar(&cfg.HelpRemote, "help-remote", "127.0.0.1:26657", "help page's remote addr")
	fs.BoolVar(&cfg.WithAnalytics, "with-analytics", cfg.WithAnalytics, "enable privacy-first analytics")
	// let's parse cli, baby
	if parseError := fs.Parse(args); parseError != nil {
		return cfg, parseError
	}
	// if asteroidName has default value, check whether <asteroidDir>/.TITLE exists
	if asteroidName == "CHANGEME" && osm.FileExists(asteroidDir+"/.TITLE") {
		s := string(osm.MustReadFile(asteroidDir + "/.TITLE"))
		s = strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(s)
		logger.Debug(fmt.Sprintf("asteroidName is %s and exists %s -> changing asteroidName to %q", asteroidName, asteroidDir+"/.TITLE", s))
		asteroidName = s
	}
	// asteroidDir must be transformed into asteroidFs
	asteroidFs = os.DirFS(asteroidDir)
	return cfg, nil
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
			Render(w, r, "views/asteroidHome.html", "views/asteroidFuncs.html")
	})
}

// this is a fallthrough handler to
// attempt serving markdown, images
// in the structure.
func HandlerAnything(logger *slog.Logger, app gotuna.App, cfg *gnoweb.Config) http.Handler {
	// Note: we read from asteroidFs (asteroidDir is not always used)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		url = html.UnescapeString(url)
		// TODO test it still works (removing leading /) with asteroidDir
		// We need to remove leading / with embed fs...
		url = strings.TrimPrefix(url, "/")
		url = filepath.Clean(url)

		if strings.Contains(url, "..") {
			app.NewTemplatingEngine().
				Set("AsteroidName", asteroidName).
				Set("Config", cfg).
				Render(w, r, "views/asteroid403.html", "views/asteroidFuncs.html")
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
				Render(w, r, "views/asteroidFuncs.html", "views/asteroidGeneric.html")
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

func StyleFsFrom(styleDir string) fs.FS {
	if styleDir == "" {
		return DefaultStyle()
	}
	return os.DirFS(styleDir)
}

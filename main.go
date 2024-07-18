package main

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
	startedAt    time.Time
	asteroidName string // read from cmdLine, or file called .TITLE at root, or "CHANGEME"
	asteroidDir  string
	bindAddr     string
)

//go:embed views/*.html
var asteroidTemplates embed.FS // composed with gnoweb's, using merged_fs

func init() {
	startedAt = time.Now()
}

func parseArgs(args []string, logger *slog.Logger) (gnoweb.Config, error) {
	cfg := gnoweb.NewDefaultConfig()
	fs := flag.NewFlagSet("gnoweb", flag.ContinueOnError)
	fs.StringVar(&asteroidDir, "asteroid-dir", "", "wiki directory location. [Mandatory!]")
	fs.StringVar(&asteroidName, "asteroid-name", "CHANGEME", "the asteroid name (website title). read from .TITLE, or CHANGEME")
	fs.StringVar(&bindAddr, "bind", "0.0.0.0:8888", "server listening address")
	fs.StringVar(&cfg.StyleDir, "style-dir", "default-style", "style directory (css, js, img). Default is to use embedded gnoweb style")
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

	return cfg, nil
}

func main() {
	zapLogger := log.NewZapConsoleLogger(os.Stdout, zapcore.DebugLevel)
	logger := log.ZapLoggerToSlog(zapLogger)

	cfg, e := parseArgs(os.Args[1:], logger)
	if e != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", e)
		os.Exit(1)
	}
	logger.Info(fmt.Sprintf("Serving %s on http://%s", asteroidDir, bindAddr))

	gnowebViews, e := fs.Sub(gnoweb.DefaultViewsFiles(), "views")
	if e != nil {
		panic("Could not find gnoweb views: " + e.Error())
	}
	makeApp := func() http.Handler {
		return gnoweb.MakeAppWithOptions(logger, cfg, gnoweb.Options{
			RootHandler:     handlerHome,
			NotFoundHandler: handlerAnything,
			ViewFS: merged_fs.NewMergedFS(
				gnowebViews,
				asteroidTemplates,
			),
		}).Router
	}

	server := &http.Server{
		Addr:              bindAddr,
		ReadHeaderTimeout: 60 * time.Second,
		Handler:           makeApp(),
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
				server.Handler = makeApp()
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
						server.Handler = makeApp()
					}
				}
			}
		}()
		watcher.AddRecursive(asteroidDir)
		if cfg.StyleDir != "" {
			watcher.AddRecursive(cfg.StyleDir)
		}
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error(fmt.Sprintf("HTTP server stopped with error: %+v\n", err))
	}
	zapLogger.Sync()
}

func handlerHome(logger *slog.Logger, app gotuna.App, cfg *gnoweb.Config) http.Handler {
	index_md := filepath.Join(asteroidDir, "index.md")
	homeContent := osm.MustReadFile(index_md)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Set("AsteroidName", asteroidName).
			Set("HomeContent", string(homeContent)).
			Set("Config", cfg).
			Render(w, r, "views/asteroidHome.html", "views/asteroidFuncs.html")
	})
}

func handlerAnything(logger *slog.Logger, app gotuna.App, cfg *gnoweb.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		url = html.UnescapeString(url)
		url = filepath.Clean(url)
		file := filepath.Join(asteroidDir, url)
		switch {
		case osm.DirExists(file) && osm.FileExists(file+"/index.md"):
			file = file + "/index.md"
		case osm.FileExists(file + ".md"):
			file = file + ".md"
		case osm.DirExists(file):
			http.Error(w, "Teapot: missing index.md", http.StatusTeapot)
			return
		case osm.FileExists(file):
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		if strings.Contains(file, "..") {
			app.NewTemplatingEngine().
				Set("AsteroidName", asteroidName).
				Set("Config", cfg).
				Render(w, r, "views/asteroid403.html", "views/asteroidFuncs.html")
		}
		// serve based on file extension
		switch {
		case strings.HasSuffix(file, ".md"):
			content := osm.MustReadFile(file)
			app.NewTemplatingEngine().
				Set("AsteroidName", asteroidName).
				Set("MainContent", string(content)).
				Set("Config", cfg).
				Render(w, r, "views/asteroidFuncs.html", "views/asteroidGeneric.html")
		case strings.HasSuffix(file, ".jpg"),
			strings.HasSuffix(file, ".jpeg"),
			strings.HasSuffix(file, ".png"),
			strings.HasSuffix(file, ".gif"):
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

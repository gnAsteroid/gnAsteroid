package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// ------------------------ copied over from cmd/gnoweb
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"go.uber.org/zap/zapcore"
	// ------------------------
	"github.com/dietsche/rfsnotify"
	"gopkg.in/fsnotify.v1"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
	"github.com/grepsuzette/gnAsteroid/static"   // for static files
)

var (
	startedAt    time.Time
	asteroidName string // read from cmdLine, or file called .TITLE at root, or "CHANGEME"
	asteroidDir  string
	bindAddr     string
)

func init() {
	startedAt = time.Now()
}

func parseArgs(args []string) (gnoweb.Config, error) {
	cfg := gnoweb.NewDefaultConfig()
	fs := flag.NewFlagSet("gnoweb", flag.ContinueOnError)
	fs.StringVar(&asteroidDir, "asteroid-dir", "", "wiki directory location. [Mandatory!]")
	fs.StringVar(&asteroidName, "asteroid-name", "CHANGEME", "the asteroid name (website title). read from .TITLE, or CHANGEME")
	fs.StringVar(&bindAddr, "bind", "0.0.0.0:8888", "server listening address")
	fs.StringVar(&cfg.StyleDir, "style-dir", "", "style directory (css, js, img). Default is to use embedded gnoweb style")
	fs.StringVar(&cfg.RemoteAddr, "remote", "127.0.0.1:26657", "remote gnoland node address")
	fs.StringVar(&cfg.CaptchaSite, "captcha-site", "", "recaptcha site key (if empty, captcha are disabled)")
	fs.StringVar(&cfg.FaucetURL, "faucet-url", "http://localhost:5050", "faucet server URL")
	fs.StringVar(&cfg.HelpChainID, "help-chainid", "dev", "help page's chainid")
	fs.StringVar(&cfg.HelpRemote, "help-remote", "127.0.0.1:26657", "help page's remote addr")
	fs.StringVar(&cfg.WithAnalytics, "with-analytics", cfg.WithAnalytics, "enable privacy-first analytics")

	// if asteroidName has default value, check if <asteroidDir>/.TITLE exists
	if asteroidName == "CHANGEME" && osm.FileExists(asteroidDir+"/.TITLE") {
		s := string(osm.MustReadFile(asteroidDir + "/.TITLE"))
		asteroidName = strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(s)
	}

	return cfg, fs.Parse(args)
}

func main() {
	cfg, e := parseArgs(os.Args[1:])
	if e != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	zapLogger := log.NewZapConsoleLogger(os.Stdout, zapcore.DebugLevel)
	logger := log.ZapLoggerToSlog(zapLogger)
	logger.Info("Serving", asteroidDir, " on ", "http://", bindAddr)

	makeApp := func() http.Handler {
		return gnoweb.MakeAppWithOptions(logger, cfg, gnoweb.Options{
			RootHandler:     handlerHome,
			NotFoundHandler: handlerAnything,
		}).Router
	}

	none := map[string]string{}
	server := &http.Server{
		Addr:              bindAddr,
		ReadHeaderTimeout: 60 * time.Second,
		Handler:           makeApp(),
	}

	// Detect SIGUSR1 -> invalidate all templates
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
						logger.Info("Reloading, modified: ", event.Name)
						server.Handler = makeApp()
					}
				}
			}
		}()
		watcher.AddRecursive(flags.asteroidDir)
		if flags.StyleDir != "" {
			watcher.AddRecursive(flags.StyleDir)
		}
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("HTTP server stopped with error: %+v\n", err)
	}
	return zapLogger.Sync()
}

func handlerHome(_, app gotuna.App) http.Handler {
	md := filepath.Join(flags.asteroidDir, "index.md")
	homeContent := osm.MustReadFile(md)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Set("AsteroidName", asteroidName).
			Set("HomeContent", string(homeContent)).
			Render(w, r, "views/home.html", "views/funcs.html")
	})
}

func handlerAnything(_, app gotuna.App) http.Handler {
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
				Set("AsteroidName", asteroidName).
				Render(w, r, "views/403.html", "views/funcs.html")
		}
		// serve the file, based of its extension
		switch {
		case strings.HasSuffix(file, ".md"):
			content := osm.MustReadFile(file)
			app.NewTemplatingEngine().
				Set("AsteroidName", asteroidName).
				Set("MainContent", string(content)).
				Render(w, r, "views/generic.html", "views/funcs.html")
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

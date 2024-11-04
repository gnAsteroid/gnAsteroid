package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dietsche/rfsnotify"
	"github.com/gnAsteroid/gnAsteroid"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"go.uber.org/zap/zapcore"
)

var bindAddr string
var themeDir string
var asteroidDir string // asteroidDir will be read and become asteroidFs

// Launch a gnAsteroid server (using gnoweb) on bindAddr
// Watch asteroid, theme dirs, SIGUSR1 for reload.
func main() {
	zapLogger := log.NewZapConsoleLogger(os.Stdout, zapcore.DebugLevel)
	logger := log.ZapLoggerToSlog(zapLogger)

	cfg, e := parseArgs(os.Args[1:], logger)
	if e != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", e)
		os.Exit(1)
	}
	logger.Info(fmt.Sprintf("Serving %s on http://%s", asteroidDir, bindAddr))
	themeFs := ThemeFsFrom(themeDir)
	server := &http.Server{
		Addr:              bindAddr,
		ReadHeaderTimeout: 60 * time.Second,
		Handler:           gnAsteroid.MakeApp(logger, cfg, themeFs),
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
				server.Handler = gnAsteroid.MakeApp(logger, cfg, themeFs)
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
					if ok {
						logger.Info("Reloading, modified: " + event.Name)
						server.Handler = gnAsteroid.MakeApp(logger, cfg, themeFs)
					}
				}
			}
		}()
		watcher.AddRecursive(asteroidDir)
		if themeDir != "" {
			watcher.AddRecursive(themeDir)
		}
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error(fmt.Sprintf("HTTP server stopped with error: %+v\n", err))
	}
	zapLogger.Sync()
}

func parseArgs(args []string, logger *slog.Logger) (gnoweb.Config, error) {
	cfg := gnoweb.NewDefaultConfig()
	flag := flag.NewFlagSet("gnoweb", flag.ContinueOnError)
	// gnAsteroid flags
	var asteroidName string
	flag.StringVar(&asteroidDir, "asteroid-dir", "", "wiki directory location. [Mandatory!]")
	flag.StringVar(&asteroidName, "asteroid-name", "CHANGEME", "the asteroid name (website title). read from .TITLE, or CHANGEME")
	flag.StringVar(&themeDir, "theme-dir", "", "theme directory (css, js, img). Default is 'themes/default.theme/'")
	flag.StringVar(&bindAddr, "bind", "0.0.0.0:8888", "server listening address")
	// gnoweb flags
	flag.StringVar(&cfg.RemoteAddr, "remote", "https://rpc.gno.land:443", "remote gnoland node address")
	flag.StringVar(&cfg.CaptchaSite, "captcha-site", "", "recaptcha site key (if empty, captcha are disabled)")
	flag.StringVar(&cfg.FaucetURL, "faucet-url", "http://localhost:5050", "faucet server URL")
	flag.StringVar(&cfg.HelpChainID, "help-chainid", "dev", "help page's chainid")
	flag.StringVar(&cfg.HelpRemote, "help-remote", "https://gno.land:443", "help page's remote addr")
	flag.BoolVar(&cfg.WithAnalytics, "with-analytics", cfg.WithAnalytics, "enable privacy-first analytics")
	// let's parse cli
	if parseError := flag.Parse(args); parseError != nil {
		return cfg, parseError
	}
	if asteroidDir == "" {
		return cfg, errors.New("-asteroid-dir is mandatory")
	} else if !osm.DirExists(asteroidDir) {
		return cfg, errors.New(asteroidDir + " is not a directory")
	} else if themeDir != "" && !osm.DirExists(themeDir) {
		return cfg, errors.New(themeDir + " is not a directory. -theme-dir must exist, if supplied.")
	}
	// if asteroidName has default value, check whether <asteroidDir>/.TITLE exists
	if asteroidName == "CHANGEME" && osm.FileExists(asteroidDir+"/.TITLE") {
		s := string(osm.MustReadFile(asteroidDir + "/.TITLE"))
		s = strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(s)
		logger.Debug(fmt.Sprintf("asteroidName is %s and exists %s -> changing asteroidName to %q", asteroidName, asteroidDir+"/.TITLE", s))
		asteroidName = s
	}
	gnAsteroid.SetAsteroidName(asteroidName)
	gnAsteroid.SetAsteroidFs(os.DirFS(asteroidDir))
	return cfg, nil
}

func ThemeFsFrom(themeDir string) fs.FS {
	if themeDir == "" {
		return os.DirFS(gnAsteroid.DefaultTheme)
	}
	return os.DirFS(themeDir)
}

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
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/grepsuzette/gnAsteroid"
	"go.uber.org/zap/zapcore"
	"gopkg.in/fsnotify.v1"
)

var bindAddr string
var styleDir string
var asteroidDir string // asteroidDir will be read and become asteroidFs

// main() for cli gnAsteroid.
// Launch a gnAsteroid server (using gnoweb) on bindAddr
// Watch asteroid, style dirs, SIGUSR1 for reload.
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
		Handler:           gnAsteroid.MakeApp(logger, cfg, styleFs),
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
				server.Handler = gnAsteroid.MakeApp(logger, cfg, styleFs)
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
						server.Handler = gnAsteroid.MakeApp(logger, cfg, styleFs)
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
	flag := flag.NewFlagSet("gnoweb", flag.ContinueOnError)
	// gnAsteroid flags
	var asteroidName string
	flag.StringVar(&asteroidDir, "asteroid-dir", "", "wiki directory location. [Mandatory!]")
	flag.StringVar(&asteroidName, "asteroid-name", "CHANGEME", "the asteroid name (website title). read from .TITLE, or CHANGEME")
	flag.StringVar(&styleDir, "style-dir", "", "style directory (css, js, img). Default is '', meaning \"use embed 'default-style/'\"")
	flag.StringVar(&bindAddr, "bind", "0.0.0.0:8888", "server listening address")
	// gnoweb flags
	flag.StringVar(&cfg.RemoteAddr, "remote", "gno.land:26657", "remote gnoland node address")
	flag.StringVar(&cfg.CaptchaSite, "captcha-site", "", "recaptcha site key (if empty, captcha are disabled)")
	flag.StringVar(&cfg.FaucetURL, "faucet-url", "http://localhost:5050", "faucet server URL")
	flag.StringVar(&cfg.HelpChainID, "help-chainid", "dev", "help page's chainid")
	flag.StringVar(&cfg.HelpRemote, "help-remote", "gno.land:26657", "help page's remote addr")
	flag.BoolVar(&cfg.WithAnalytics, "with-analytics", cfg.WithAnalytics, "enable privacy-first analytics")
	// let's parse cli
	if parseError := flag.Parse(args); parseError != nil {
		return cfg, parseError
	}
	if asteroidDir == "" {
		return cfg, errors.New("-asteroid-dir is mandatory")
	} else if !osm.DirExists(asteroidDir) {
		return cfg, errors.New(asteroidDir + " is not a directory")
	} else if styleDir != "" && !osm.DirExists(styleDir) {
		return cfg, errors.New(styleDir + " is not a directory. -style-dir must exist, if supplied.")
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

func StyleFsFrom(styleDir string) fs.FS {
	if styleDir == "" {
		return gnAsteroid.DefaultStyle()
	}
	return os.DirFS(styleDir)
}

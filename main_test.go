package gnAsteroid

import (
	"embed"
	"io/fs"
	"log/slog"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed neptune
var neptune embed.FS

func TestAsteroid(t *testing.T) {
	neptuneFs, _ := fs.Sub(neptune, "neptune")

	if nil == HandleAsteroid(
		neptuneFs, DefaultStyle(), "neptune as an asteroid",
		gnoweb.Config{
			RemoteAddr:  "gno.land:26657",
			HelpChainID: "portal-loop",
			HelpRemote:  "gno.land:26657",
		},
	) {
		t.Error("nil handler")
	}

	// TODO add serving test
}

func TestParseArgs(t *testing.T) {
	const (
		name       = "my asteroid"
		exampleDir = "example"
		bind       = "57182"
		styleDir   = "style" // a fake style dir, it's not checked here
	)
	_, e := parseArgs([]string{
		"-asteroid-dir", exampleDir,
		"-asteroid-name", name,
		"-bind", bind,
		"-style-dir", styleDir,
	}, slog.Default())
	if e != nil {
		t.Errorf("Could not parse args, %v", e)
		t.FailNow()
	}
	require.Equal(t, asteroidName, name)
	require.Equal(t, asteroidDir, exampleDir)
	require.Equal(t, bindAddr, bind)
	require.Equal(t, styleDir, styleDir)
}

// an asteroid read from -asteroid-dir
// (the other case is using HandleAsteroid, using embed)
func TestAsteroidFromDir(t *testing.T) {
	// TODO add this test
}

// TODO
// favicon.ico gives Content-Type image/x-icon, 200 OK

package gnAsteroid

import (
	"log/slog"
	"testing"

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	const (
		name       = "Golf November Oscar"
		exampleDir = "example"
		bind       = "57182"
		styleDir   = "style" // it's a fake style dir, it's not checked here
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

// TODO
// favicon.ico gives Content-Type image/x-icon, 200 OK

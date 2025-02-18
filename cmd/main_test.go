package main

import (
	"testing"

	"log/slog"

	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	const (
		name       = "my asteroid"
		exampleDir = "../example"
		bind       = "57182"
		themeDir   = "../themes/raw.theme"
	)
	_, e := parseArgs([]string{
		"-asteroid-dir", exampleDir,
		"-asteroid-name", name,
		"-bind", bind,
		"-theme-dir", themeDir,
	}, slog.Default())
	if e != nil {
		t.Errorf("Could not parse args, %v", e)
		t.FailNow()
	}
	// require.Equal(t, asteroidName, name)
	require.Equal(t, asteroidDir, exampleDir)
	require.Equal(t, bindAddr, bind)
}

// an asteroid read from -asteroid-dir
// (the other case is using HandleAsteroid, using embed)
func TestAsteroidFromDir(t *testing.T) {
	// TODO add this test
}

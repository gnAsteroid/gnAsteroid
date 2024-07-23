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
		styleDir   = "../default-style" // a fake style dir, it's not checked here
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
	// require.Equal(t, asteroidName, name)
	require.Equal(t, asteroidDir, exampleDir)
	require.Equal(t, bindAddr, bind)
	require.Equal(t, styleDir, styleDir)
}

// an asteroid read from -asteroid-dir
// (the other case is using HandleAsteroid, using embed)
func TestAsteroidFromDir(t *testing.T) {
	// TODO add this test
}

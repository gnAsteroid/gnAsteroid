package gnAsteroid
import (
	"embed"
	"io/fs"
	"os"
	"testing"

	"github.com/gnAsteroid/gno/gno.land/pkg/gnoweb"
    "github.com/stretchr/testify/assert"
)

//go:embed example
var neptune embed.FS

func TestAsteroid(t *testing.T) {
	neptuneFs, _ := fs.Sub(neptune, "example")

	if nil == HandleAsteroid(
		neptuneFs, os.DirFS(DefaultTheme), "neptune as an asteroid",
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

func TestExtractFrontMatter(t *testing.T) {
    {
        example := `---
title: Qwerty Uiop!@#$%^&*()
tags: [great, many, many, tags]
date: 2028932312341234237417234192374
---
Ola,
amigo`
        md, kv := ExtractFrontMatter(example) 
        assert.Equal(t, kv["title"], "Qwerty Uiop!@#$%^&*()")
        assert.Equal(t, kv["tags"], "[great, many, many, tags]")
        assert.Equal(t, kv["date"], "2028932312341234237417234192374")
        assert.Equal(t, md, "Ola,\namigo\n")
    }
    {
        md, kv := ExtractFrontMatter("---title: What Is The Matrix---\nActual article") 
        assert.Equal(t, kv["title"], "What Is The Matrix")
        assert.Equal(t, md, "Actual article")
    }
}

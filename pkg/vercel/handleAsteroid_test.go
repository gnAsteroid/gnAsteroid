package vercel

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/grepsuzette/gnAsteroid"
)

//go:embed neptune
var neptune embed.FS

func TestAsteroid(t *testing.T) {
	// cope with the fact embed do not merely have their
	// location as root by default. Don't put a final "/".
	neptuneFs, _ := fs.Sub(neptune, "neptune")

	// this test mostly makes sure an index.md is readable
	// in the asteroid root.
	if nil == HandleAsteroid(
		neptuneFs, gnAsteroid.DefaultStyle(), "neptune as an asteroid",
		gnoweb.Config{
			RemoteAddr:  "gno.land:26657",
			HelpChainID: "portal-loop",
			HelpRemote:  "gno.land:26657",
		},
	) {
		t.Error("nil handler")
	}
}

package gnAsteroid

import (
	"embed"
	"io/fs"
	"os"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
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

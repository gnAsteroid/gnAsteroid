package bcver

import (
	types "github.com/gnolang/gno/pkgs/bft/types/version"
)

const Version = "v1.0.0-rc.0"

func init() {
	if types.BlockVersion != "v1.0.0-rc.0" {
		panic("bump Version")
	}
}

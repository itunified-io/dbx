// dbxcli is the interactive CLI for the dbx database lifecycle management framework.
package main

import (
	"fmt"
	"os"

	"github.com/itunified-io/dbx/cmd/dbxcli/root"
)

var version = "dev"

func main() {
	if err := root.New(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

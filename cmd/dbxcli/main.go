// dbxcli is the interactive CLI for the dbx database lifecycle management framework.
package main

import (
	"fmt"
	"os"

	"github.com/itunified-io/dbx/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("dbxcli %s\n", version.Version)
		return
	}
	fmt.Fprintln(os.Stderr, "dbxcli: not yet implemented")
	os.Exit(1)
}

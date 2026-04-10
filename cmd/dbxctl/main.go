// dbxctl is the service controller for the dbx database lifecycle management framework.
package main

import (
	"fmt"
	"os"

	"github.com/itunified-io/dbx/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("dbxctl %s\n", version.Version)
		return
	}
	fmt.Fprintln(os.Stderr, "dbxctl: not yet implemented")
	os.Exit(1)
}

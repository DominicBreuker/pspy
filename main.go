package main

import (
	"fmt"

	"github.com/dominicbreuker/pspy/cmd"
)

var version string
var commit string

func main() {
	fmt.Printf("pspy - version: %s - Commit SHA: %s\n", version, commit)
	cmd.Execute()
}

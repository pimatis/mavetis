package main

import (
	"os"

	"github.com/Pimatis/mavetis/src/cli"
)

func main() {
	os.Exit(cli.Execute(os.Args[1:]))
}

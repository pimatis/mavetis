package cli

import (
	"errors"
	"flag"
	"os"

	updater "github.com/Pimatis/mavetis/src/update"
)

var updateRun = updater.Run

func runUpdate(arguments []string) int {
	spec, err := parseUpdate(arguments)
	if err != nil {
		return fail(err)
	}
	if err := updateRun(spec); err != nil {
		return fail(err)
	}
	return 0
}

func parseUpdate(arguments []string) (updater.Spec, error) {
	spec := updater.Spec{}
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.BoolVar(&spec.Check, "check", false, "Check whether a newer GitHub release is available")
	if err := flags.Parse(arguments); err != nil {
		return spec, err
	}
	if len(flags.Args()) != 0 {
		return spec, errors.New("unexpected positional arguments")
	}
	return spec, nil
}

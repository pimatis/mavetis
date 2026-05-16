package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/Pimatis/mavetis/src/git"
	"github.com/Pimatis/mavetis/src/hook"
	"github.com/Pimatis/mavetis/src/model"
)

func Execute(arguments []string) int {
	if len(arguments) == 0 {
		usage()
		return 0
	}
	command := arguments[0]
	if command == "help" || command == "-h" || command == "--help" {
		usage()
		return 0
	}
	if command == "review" {
		return runReview(arguments[1:], false)
	}
	if command == "ci" {
		return runReview(arguments[1:], true)
	}
	if command == "hooks" {
		return runHooks(arguments[1:])
	}
	if command == "rules" {
		return runRules(arguments[1:])
	}
	if command == "explain" {
		return runExplain(arguments[1:])
	}
	if command == "update" {
		return runUpdate(arguments[1:])
	}
	if command == "shell" {
		return runShell(arguments[1:])
	}
	if command == "init" {
		return runInit(arguments[1:])
	}
	if command == "baseline" {
		return runBaseline(arguments[1:])
	}
	if command == "secrets" {
		return runSecrets(arguments[1:])
	}
	if command == "version" || command == "-v" || command == "--version" {
		fmt.Printf("%s %s\n", model.Name, model.Version)
		return 0
	}
	usage()
	return 1
}

func runHooks(arguments []string) int {
	if len(arguments) == 0 {
		return fail(errors.New("hooks command requires install or uninstall"))
	}
	root, err := git.Root()
	if err != nil {
		return fail(err)
	}
	if arguments[0] == "install" {
		if err := hook.Install(root); err != nil {
			return fail(err)
		}
		fmt.Println("hooks installed")
		return 0
	}
	if arguments[0] == "uninstall" {
		if err := hook.Uninstall(root); err != nil {
			return fail(err)
		}
		fmt.Println("hooks removed")
		return 0
	}
	return fail(errors.New("hooks command requires install or uninstall"))
}

func fail(err error) int {
	fmt.Fprintln(os.Stderr, err.Error())
	return 2
}

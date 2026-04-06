package cli

import (
	"errors"
	"fmt"
)

func runShell(arguments []string) int {
	if len(arguments) == 0 {
		return fail(errors.New("shell command requires init"))
	}
	if arguments[0] != "init" {
		return fail(errors.New("shell command requires init"))
	}
	if len(arguments) != 2 {
		return fail(errors.New("shell init requires a shell name"))
	}
	script, err := shellInitScript(arguments[1])
	if err != nil {
		return fail(err)
	}
	fmt.Print(script)
	return 0
}

func shellInitScript(shell string) (string, error) {
	if shell == "zsh" {
		return "unalias mavetis 2>/dev/null || true\nalias mavetis='noglob mavetis'\n", nil
	}
	return "", errors.New("shell init supports only zsh")
}

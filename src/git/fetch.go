package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

func Review(spec model.Review) (string, model.DiffMeta, error) {
	meta := model.DiffMeta{}
	meta.Base = spec.Base
	meta.Head = spec.Head
	if spec.Staged {
		meta.Mode = "staged"
		output, err := run("diff", "--cached", "--no-ext-diff", "--no-color", "--unified=3")
		return output, meta, err
	}
	if spec.Base != "" {
		meta.Mode = "branch"
		head := spec.Head
		if head == "" {
			head = "HEAD"
			meta.Head = head
		}
		output, err := run("diff", "--no-ext-diff", "--no-color", "--unified=3", spec.Base+"..."+head)
		return output, meta, err
	}
	meta.Mode = "worktree"
	output, err := run("diff", "--no-ext-diff", "--no-color", "--unified=3")
	return output, meta, err
}

func Root() (string, error) {
	output, err := run("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func run(args ...string) (string, error) {
	return runIn("", args...)
}

func runIn(cwd string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	if cwd != "" {
		command.Dir = cwd
	}
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err == nil {
		return stdout.String(), nil
	}
	message := strings.TrimSpace(stderr.String())
	if message == "" {
		message = err.Error()
	}
	return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
}

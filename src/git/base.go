package git

import "strings"

func DefaultBase(root string) string {
	if root != "" {
		branch := remoteBase(root)
		if branch != "" {
			return branch
		}
	}
	for _, item := range []string{"main", "master", "trunk", "develop"} {
		if exists(root, item) {
			return item
		}
	}
	return "main"
}

func remoteBase(root string) string {
	output, err := runIn(root, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(output)
	if text == "" {
		return ""
	}
	parts := strings.Split(text, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func exists(root string, branch string) bool {
	_, err := runIn(root, "rev-parse", "--verify", branch)
	return err == nil
}

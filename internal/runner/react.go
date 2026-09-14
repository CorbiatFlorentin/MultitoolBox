package runner

import (
	"os"
	"strings"
)

func init() {
	Register(reactRunner{})
}

type reactRunner struct{}

func (reactRunner) Name() string { return "react" }

func (reactRunner) Detect(dir string) bool {
	data, err := os.ReadFile(dir + string(os.PathSeparator) + "package.json")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), `"react"`)
}

func (reactRunner) Test(dir string) error {
	return run(dir, "npm", "test", "--silent")
}

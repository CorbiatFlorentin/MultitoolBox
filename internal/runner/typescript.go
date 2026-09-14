package runner

import (
	"fmt"
	"path/filepath"
)

func init() {
	Register(typescriptRunner{})
}

type typescriptRunner struct{}

func (typescriptRunner) Name() string { return "typescript" }

func (typescriptRunner) Detect(dir string) bool {
	return fileExists(dir, "tsconfig.json")
}

func (typescriptRunner) Test(dir string) error {
	// On invoke directement le script tsc du paquet "typescript" du projet
	// (via node) plutôt que "npx tsc" : npx peut résoudre un paquet npm
	// homonyme sans rapport (ex. le "tsc" squatteur) si "typescript" n'est
	// pas une dépendance du projet, ou si ce squatteur est déjà en cache npx.
	tscScript := filepath.Join("node_modules", "typescript", "bin", "tsc")
	if !fileExists(dir, tscScript) {
		return fmt.Errorf("typescript n'est pas installé dans %s (npm install --save-dev typescript)", dir)
	}

	if err := run(dir, "node", tscScript, "--noEmit"); err != nil {
		return err
	}
	return run(dir, "npm", "test", "--silent")
}

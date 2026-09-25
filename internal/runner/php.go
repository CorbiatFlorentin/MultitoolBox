package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func init() {
	Register(phpRunner{})
}

type phpRunner struct{}

func (phpRunner) Name() string { return "php" }

// Detect ne se contente pas d'un composer.json : une lib PHP sans aucun test
// ne doit pas être lancée par "multitest test all". Il faut un signal de
// tests (config phpunit/pest, Laravel, script composer "test", dépendance de
// test ou dossier tests/).
func (phpRunner) Detect(dir string) bool {
	for _, f := range []string{"phpunit.xml", "phpunit.xml.dist", "artisan", filepath.Join("tests", "Pest.php")} {
		if fileExists(dir, f) {
			return true
		}
	}
	if !fileExists(dir, "composer.json") {
		return false
	}
	if fileExists(dir, "tests") {
		return true
	}
	cj, err := readComposerJSON(dir)
	if err != nil {
		return false
	}
	return cj.hasTestScript() || cj.hasDevDep("phpunit/phpunit") || cj.hasDevDep("pestphp/pest")
}

// Test choisit le lanceur par ordre de priorité : Laravel (php artisan test),
// Pest, PHPUnit, puis le script composer "test". Les binaires de vendor/bin
// sont lancés via "php" : ce sont des scripts PHP sans extension, que
// Windows ne sait pas exécuter directement.
func (phpRunner) Test(dir string) error {
	hasVendor := fileExists(dir, "vendor")
	pest := filepath.Join("vendor", "bin", "pest")
	phpunit := filepath.Join("vendor", "bin", "phpunit")

	switch {
	case hasVendor && fileExists(dir, "artisan"):
		return run(dir, "php", "artisan", "test")
	case fileExists(dir, pest):
		return run(dir, "php", pest)
	case fileExists(dir, phpunit):
		return run(dir, "php", phpunit)
	}

	if !fileExists(dir, "composer.json") {
		return fmt.Errorf("aucun lanceur de tests PHP trouvé dans %s (ni vendor/bin/phpunit, ni composer.json)", dir)
	}
	cj, err := readComposerJSON(dir)
	if err != nil {
		return fmt.Errorf("composer.json illisible dans %s: %w", dir, err)
	}
	if cj.hasTestScript() {
		return run(dir, "composer", "test")
	}
	if !hasVendor {
		return fmt.Errorf("dépendances PHP absentes dans %s : lance d'abord \"composer install\"", dir)
	}
	return fmt.Errorf("aucun lanceur de tests PHP trouvé dans %s : installe phpunit/pest ou ajoute un script \"test\" dans composer.json", dir)
}

type composerJSON struct {
	Scripts    map[string]json.RawMessage `json:"scripts"`
	RequireDev map[string]string          `json:"require-dev"`
}

func readComposerJSON(dir string) (composerJSON, error) {
	var cj composerJSON
	data, err := os.ReadFile(filepath.Join(dir, "composer.json"))
	if err != nil {
		return cj, err
	}
	err = json.Unmarshal(data, &cj)
	return cj, err
}

func (c composerJSON) hasTestScript() bool {
	_, ok := c.Scripts["test"]
	return ok
}

func (c composerJSON) hasDevDep(name string) bool {
	_, ok := c.RequireDev[name]
	return ok
}

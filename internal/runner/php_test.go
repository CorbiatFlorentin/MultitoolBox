package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// touch crée un fichier vide (et ses dossiers parents) sous dir.
func touch(t *testing.T, dir string, parts ...string) {
	t.Helper()
	path := filepath.Join(append([]string{dir}, parts...)...)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPhpDetect(t *testing.T) {
	var r phpRunner

	detected := []struct {
		name  string
		setup func(t *testing.T, dir string)
	}{
		{"phpunit.xml", func(t *testing.T, dir string) { touch(t, dir, "phpunit.xml") }},
		{"phpunit.xml.dist", func(t *testing.T, dir string) { touch(t, dir, "phpunit.xml.dist") }},
		{"Laravel (artisan)", func(t *testing.T, dir string) { touch(t, dir, "artisan") }},
		{"Pest (tests/Pest.php)", func(t *testing.T, dir string) { touch(t, dir, "tests", "Pest.php") }},
		{"composer.json + dossier tests/", func(t *testing.T, dir string) {
			writeFile(t, dir, "composer.json", `{}`)
			touch(t, dir, "tests", "ExampleTest.php")
		}},
		{"composer.json avec script test", func(t *testing.T, dir string) {
			writeFile(t, dir, "composer.json", `{"scripts": {"test": "phpunit"}}`)
		}},
		{"composer.json avec phpunit en require-dev", func(t *testing.T, dir string) {
			writeFile(t, dir, "composer.json", `{"require-dev": {"phpunit/phpunit": "^11"}}`)
		}},
		{"composer.json avec pest en require-dev", func(t *testing.T, dir string) {
			writeFile(t, dir, "composer.json", `{"require-dev": {"pestphp/pest": "^3"}}`)
		}},
	}
	for _, c := range detected {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			c.setup(t, dir)
			if !r.Detect(dir) {
				t.Errorf("Detect() = false, want true (%s)", c.name)
			}
		})
	}

	notDetected := []struct {
		name  string
		setup func(t *testing.T, dir string)
	}{
		{"dossier vide", func(t *testing.T, dir string) {}},
		{"lib sans aucun test", func(t *testing.T, dir string) {
			writeFile(t, dir, "composer.json", `{"require": {"php": ">=8.1"}}`)
		}},
		{"composer.json invalide", func(t *testing.T, dir string) {
			writeFile(t, dir, "composer.json", `{pas du json`)
		}},
	}
	for _, c := range notDetected {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			c.setup(t, dir)
			if r.Detect(dir) {
				t.Errorf("Detect() = true, want false (%s)", c.name)
			}
		})
	}
}

func TestPhpTestPriority(t *testing.T) {
	pest := filepath.Join("vendor", "bin", "pest")
	phpunit := filepath.Join("vendor", "bin", "phpunit")

	cases := []struct {
		name     string
		files    [][]string
		wantArgs string
	}{
		{"Laravel prioritaire sur pest et phpunit",
			[][]string{{"artisan"}, {"vendor", "bin", "pest"}, {"vendor", "bin", "phpunit"}},
			"artisan test"},
		{"pest prioritaire sur phpunit",
			[][]string{{"vendor", "bin", "pest"}, {"vendor", "bin", "phpunit"}},
			pest},
		{"phpunit seul",
			[][]string{{"vendor", "bin", "phpunit"}},
			phpunit},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range c.files {
				touch(t, dir, f...)
			}
			out := withFakeCommand(t, "php", 0)
			var r phpRunner
			if err := r.Test(dir); err != nil {
				t.Fatalf("Test() a retourné une erreur inattendue: %v", err)
			}
			if got := readOut(t, out); got != c.wantArgs {
				t.Errorf("arguments passés à php = %q, want %q", got, c.wantArgs)
			}
		})
	}
}

// Sous Windows, vendor/bin/phpunit est un script sans extension : il doit
// être lancé via "php" et jamais exécuté directement.
func TestPhpTestRunsVendorBinThroughPhp(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "vendor", "bin", "phpunit")
	phpOut := withFakeCommand(t, "php", 0)

	var r phpRunner
	if err := r.Test(dir); err != nil {
		t.Fatalf("Test() a retourné une erreur inattendue: %v", err)
	}
	if _, err := os.Stat(phpOut); err != nil {
		t.Fatal("php n'a pas été appelé pour lancer vendor/bin/phpunit")
	}
}

func TestPhpTestArtisanWithoutVendorNeedsInstall(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "artisan")
	writeFile(t, dir, "composer.json", `{"require-dev": {"phpunit/phpunit": "^11"}}`)

	var r phpRunner
	err := r.Test(dir)
	if err == nil || !strings.Contains(err.Error(), "composer install") {
		t.Errorf("Test() = %v, want une erreur conseillant \"composer install\"", err)
	}
}

func TestPhpTestFallsBackToComposerScript(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.json", `{"scripts": {"test": ["@php vendor/bin/phpunit"]}}`)
	out := withFakeCommand(t, "composer", 0)

	var r phpRunner
	if err := r.Test(dir); err != nil {
		t.Fatalf("Test() a retourné une erreur inattendue: %v", err)
	}
	if got, want := readOut(t, out), "test"; got != want {
		t.Errorf("arguments passés à composer = %q, want %q", got, want)
	}
}

func TestPhpTestPropagatesFailure(t *testing.T) {
	t.Run("phpunit échoue", func(t *testing.T) {
		dir := t.TempDir()
		touch(t, dir, "vendor", "bin", "phpunit")
		withFakeCommand(t, "php", 1)
		var r phpRunner
		if err := r.Test(dir); err == nil {
			t.Error("Test() = nil, want une erreur quand phpunit échoue")
		}
	})

	t.Run("composer test échoue", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "composer.json", `{"scripts": {"test": "phpunit"}}`)
		withFakeCommand(t, "composer", 1)
		var r phpRunner
		if err := r.Test(dir); err == nil {
			t.Error("Test() = nil, want une erreur quand composer test échoue")
		}
	})
}

func TestPhpTestHelpfulErrors(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(t *testing.T, dir string)
		wantMsg string
	}{
		{"dossier vide", func(t *testing.T, dir string) {}, "aucun lanceur"},
		{"vendor/ absent", func(t *testing.T, dir string) {
			writeFile(t, dir, "composer.json", `{"require-dev": {"phpunit/phpunit": "^11"}}`)
		}, "composer install"},
		{"vendor/ présent mais ni phpunit ni script test", func(t *testing.T, dir string) {
			writeFile(t, dir, "composer.json", `{}`)
			touch(t, dir, "vendor", "autoload.php")
		}, "script \"test\""},
		{"composer.json invalide", func(t *testing.T, dir string) {
			writeFile(t, dir, "composer.json", `{pas du json`)
		}, "illisible"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			c.setup(t, dir)
			// Filets de sécurité : ces cas ne doivent lancer aucune commande.
			phpOut := withFakeCommand(t, "php", 0)
			composerOut := withFakeCommand(t, "composer", 0)

			var r phpRunner
			err := r.Test(dir)
			if err == nil || !strings.Contains(err.Error(), c.wantMsg) {
				t.Errorf("Test() = %v, want une erreur contenant %q", err, c.wantMsg)
			}
			for _, out := range []string{phpOut, composerOut} {
				if _, err := os.Stat(out); err == nil {
					t.Errorf("%s a été exécuté alors qu'aucun lanceur n'est valide", filepath.Base(out))
				}
			}
		})
	}
}

// TestPhpDemoIntegration lance les vrais tests PHPUnit de examples/demo-php.
// Ignoré si php n'est pas installé ou si "composer install" n'a pas été fait
// dans l'exemple (la CI le fait).
func TestPhpDemoIntegration(t *testing.T) {
	demo, err := filepath.Abs(filepath.Join("..", "..", "examples", "demo-php"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("php"); err != nil {
		t.Skip("php absent du PATH")
	}
	if !fileExists(demo, filepath.Join("vendor", "bin", "phpunit")) {
		t.Skip("examples/demo-php/vendor absent : lance \"composer install\" dans l'exemple")
	}

	var r phpRunner
	if !r.Detect(demo) {
		t.Fatal("Detect() = false sur examples/demo-php")
	}
	if err := r.Test(demo); err != nil {
		t.Fatalf("les tests PHPUnit de la démo échouent: %v", err)
	}
}

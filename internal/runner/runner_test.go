package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile(%s): %v", name, err)
	}
}

func TestPythonDetect(t *testing.T) {
	cases := []struct {
		name string
		file string
	}{
		{"requirements.txt", "requirements.txt"},
		{"pyproject.toml", "pyproject.toml"},
		{"setup.py", "setup.py"},
	}

	var r pythonRunner

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, c.file, "")
			if !r.Detect(dir) {
				t.Errorf("Detect() = false, want true avec %s présent", c.file)
			}
		})
	}

	t.Run("dossier vide", func(t *testing.T) {
		dir := t.TempDir()
		if r.Detect(dir) {
			t.Error("Detect() = true, want false sur un dossier vide")
		}
	})
}

func TestTypescriptDetect(t *testing.T) {
	var r typescriptRunner

	t.Run("avec tsconfig.json", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "tsconfig.json", "{}")
		if !r.Detect(dir) {
			t.Error("Detect() = false, want true avec tsconfig.json présent")
		}
	})

	t.Run("dossier vide", func(t *testing.T) {
		dir := t.TempDir()
		if r.Detect(dir) {
			t.Error("Detect() = true, want false sur un dossier vide")
		}
	})
}

func TestReactDetect(t *testing.T) {
	var r reactRunner

	t.Run("package.json avec react", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "package.json", `{"dependencies": {"react": "^18.0.0"}}`)
		if !r.Detect(dir) {
			t.Error("Detect() = false, want true quand package.json contient react")
		}
	})

	t.Run("package.json sans react", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "package.json", `{"dependencies": {"vue": "^3.0.0"}}`)
		if r.Detect(dir) {
			t.Error("Detect() = true, want false quand package.json ne contient pas react")
		}
	})

	t.Run("pas de package.json", func(t *testing.T) {
		dir := t.TempDir()
		if r.Detect(dir) {
			t.Error("Detect() = true, want false sans package.json")
		}
	})
}

func TestPythonTestInvokesPytest(t *testing.T) {
	out := withFakeCommand(t, "python", 0)
	var r pythonRunner

	if err := r.Test(t.TempDir()); err != nil {
		t.Fatalf("Test() a retourné une erreur inattendue: %v", err)
	}
	if got, want := readOut(t, out), "-m pytest"; got != want {
		t.Errorf("arguments passés à python = %q, want %q", got, want)
	}
}

func TestPythonTestPropagatesFailure(t *testing.T) {
	withFakeCommand(t, "python", 1)
	var r pythonRunner

	if err := r.Test(t.TempDir()); err == nil {
		t.Error("Test() = nil, want une erreur quand pytest échoue")
	}
}

func TestReactTestInvokesNpmTest(t *testing.T) {
	out := withFakeCommand(t, "npm", 0)
	var r reactRunner

	if err := r.Test(t.TempDir()); err != nil {
		t.Fatalf("Test() a retourné une erreur inattendue: %v", err)
	}
	if got, want := readOut(t, out), "test --silent"; got != want {
		t.Errorf("arguments passés à npm = %q, want %q", got, want)
	}
}

func TestReactTestPropagatesFailure(t *testing.T) {
	withFakeCommand(t, "npm", 1)
	var r reactRunner

	if err := r.Test(t.TempDir()); err == nil {
		t.Error("Test() = nil, want une erreur quand npm test échoue")
	}
}

func TestTypescriptTestMissingTsc(t *testing.T) {
	dir := t.TempDir()
	var r typescriptRunner

	err := r.Test(dir)
	if err == nil {
		t.Fatal("Test() = nil, want une erreur quand typescript n'est pas installé")
	}
	if !strings.Contains(err.Error(), "n'est pas installé") {
		t.Errorf("message d'erreur = %q, want qu'il mentionne l'absence d'installation", err.Error())
	}
}

func TestTypescriptTestRunsTscThenNpmTest(t *testing.T) {
	dir := t.TempDir()
	tscDir := filepath.Join(dir, "node_modules", "typescript", "bin")
	if err := os.MkdirAll(tscDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, tscDir, "tsc", "")

	nodeOut := withFakeCommand(t, "node", 0)
	npmOut := withFakeCommand(t, "npm", 0)

	var r typescriptRunner
	if err := r.Test(dir); err != nil {
		t.Fatalf("Test() a retourné une erreur inattendue: %v", err)
	}

	wantSuffix := filepath.Join("node_modules", "typescript", "bin", "tsc") + " --noEmit"
	if got := readOut(t, nodeOut); !strings.HasSuffix(got, wantSuffix) {
		t.Errorf("arguments passés à node = %q, want un suffixe %q", got, wantSuffix)
	}
	if got, want := readOut(t, npmOut), "test --silent"; got != want {
		t.Errorf("arguments passés à npm = %q, want %q", got, want)
	}
}

func TestTypescriptTestStopsIfTscFails(t *testing.T) {
	dir := t.TempDir()
	tscDir := filepath.Join(dir, "node_modules", "typescript", "bin")
	if err := os.MkdirAll(tscDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, tscDir, "tsc", "")

	withFakeCommand(t, "node", 1) // tsc --noEmit échoue
	npmOut := withFakeCommand(t, "npm", 0)

	var r typescriptRunner
	if err := r.Test(dir); err == nil {
		t.Error("Test() = nil, want une erreur quand tsc échoue")
	}
	if _, err := os.Stat(npmOut); err == nil {
		t.Error("npm test a été appelé alors que tsc a déjà échoué")
	}
}

func TestRegistry(t *testing.T) {
	for _, name := range []string{"python", "react", "php", "typescript", "http"} {
		t.Run(name, func(t *testing.T) {
			r, ok := Get(name)
			if !ok {
				t.Fatalf("Get(%q) = ok:false, want ok:true (enregistré via init)", name)
			}
			if r.Name() != name {
				t.Errorf("Name() = %q, want %q", r.Name(), name)
			}
		})
	}

	t.Run("langage inconnu", func(t *testing.T) {
		if _, ok := Get("cobol"); ok {
			t.Error("Get(\"cobol\") = ok:true, want ok:false")
		}
	})
}

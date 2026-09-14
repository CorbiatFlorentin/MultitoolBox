package runner

import (
	"os"
	"path/filepath"
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

func TestPhpDetect(t *testing.T) {
	cases := []string{"composer.json", "phpunit.xml", "phpunit.xml.dist"}

	var r phpRunner

	for _, file := range cases {
		t.Run(file, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, file, "")
			if !r.Detect(dir) {
				t.Errorf("Detect() = false, want true avec %s présent", file)
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

func TestRegistry(t *testing.T) {
	for _, name := range []string{"python", "react", "php", "typescript"} {
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

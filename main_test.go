package main

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"multitest/internal/runner"
)

// fakeRunner est un Runner factice pour tester la logique de dispatch de
// runTest sans dépendre des vrais langages ni d'exécutables externes.
//
// Le registre de runner.Register est global et partagé entre tous les tests
// du binaire : Detect() ne renvoie donc true que si un fichier marqueur
// "<name>.marker" existe dans dir, pour qu'un fakeRunner enregistré par un
// autre test ne soit jamais détecté dans un dossier qu'il ne connaît pas.
type fakeRunner struct {
	name    string
	testErr error
	calls   *int
}

func (f fakeRunner) Name() string { return f.name }

func (f fakeRunner) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, f.name+".marker"))
	return err == nil
}

func (f fakeRunner) Test(dir string) error {
	if f.calls != nil {
		*f.calls++
	}
	return f.testErr
}

func markDetected(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".marker"), nil, 0o644); err != nil {
		t.Fatalf("création du marqueur pour %s: %v", name, err)
	}
}

func TestRunTestUnknownLanguage(t *testing.T) {
	if err := runTest("langage-totalement-inconnu-xyz", t.TempDir()); err == nil {
		t.Error("runTest() = nil, want une erreur pour un langage inconnu")
	}
}

func TestRunTestSpecificLanguageSuccess(t *testing.T) {
	calls := 0
	runner.Register(fakeRunner{name: "faux-langage-ok", calls: &calls})

	if err := runTest("faux-langage-ok", t.TempDir()); err != nil {
		t.Fatalf("runTest() a retourné une erreur inattendue: %v", err)
	}
	if calls != 1 {
		t.Errorf("Test() appelé %d fois, want 1", calls)
	}
}

func TestRunTestSpecificLanguagePropagatesError(t *testing.T) {
	wantErr := errors.New("échec des tests")
	runner.Register(fakeRunner{name: "faux-langage-echec", testErr: wantErr})

	if err := runTest("faux-langage-echec", t.TempDir()); err == nil {
		t.Error("runTest() = nil, want une erreur propagée depuis Test()")
	}
}

func TestRunTestAllRunsOnlyDetectedRunners(t *testing.T) {
	dir := t.TempDir()
	detectedCalls, skippedCalls := 0, 0
	runner.Register(fakeRunner{name: "faux-detecte", calls: &detectedCalls})
	runner.Register(fakeRunner{name: "faux-non-detecte", calls: &skippedCalls})
	markDetected(t, dir, "faux-detecte")

	if err := runTest("all", dir); err != nil {
		t.Fatalf("runTest(\"all\") a retourné une erreur inattendue: %v", err)
	}
	if detectedCalls != 1 {
		t.Errorf("runner détecté appelé %d fois, want 1", detectedCalls)
	}
	if skippedCalls != 0 {
		t.Errorf("runner non détecté appelé %d fois, want 0", skippedCalls)
	}
}

func TestRunTestAllStopsOnError(t *testing.T) {
	dir := t.TempDir()
	runner.Register(fakeRunner{name: "faux-en-echec-all", testErr: errors.New("boom")})
	markDetected(t, dir, "faux-en-echec-all")

	if err := runTest("all", dir); err == nil {
		t.Error("runTest(\"all\") = nil, want une erreur propagée depuis un runner détecté")
	}
}

func TestRunTestAllNoProjectDetected(t *testing.T) {
	// Dossier vide : aucun runner, réel ou factice, ne doit s'y détecter.
	if err := runTest("all", t.TempDir()); err != nil {
		t.Errorf("runTest(\"all\") sur un dossier vide = %v, want nil", err)
	}
}

func TestRunTestMissingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "n-existe-pas")
	for _, target := range []string{"all", "python"} {
		t.Run(target, func(t *testing.T) {
			err := runTest(target, missing)
			if err == nil || !strings.Contains(err.Error(), "dossier introuvable") {
				t.Errorf("runTest(%q, dossier absent) = %v, want \"dossier introuvable\"", target, err)
			}
		})
	}
}

func TestRunTestPathIsAFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "fichier.txt")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runTest("all", file); err == nil {
		t.Error("runTest() = nil, want une erreur quand le chemin est un fichier")
	}
}

func TestParseTestArgs(t *testing.T) {
	cases := []struct {
		name        string
		args        []string
		wantTarget  string
		wantDir     string
		wantTimeout time.Duration
		wantErr     bool
	}{
		{"langage seul", []string{"php"}, "php", ".", 0, false},
		{"langage et chemin", []string{"all", "examples"}, "all", "examples", 0, false},
		{"avec --timeout", []string{"--timeout", "90s", "python", "src"}, "python", "src", 90 * time.Second, false},
		{"sans argument", nil, "", "", 0, true},
		{"trop d'arguments", []string{"php", "a", "b"}, "", "", 0, true},
		{"--timeout invalide", []string{"--timeout", "bientôt", "php"}, "", "", 0, true},
		{"option inconnue", []string{"--verbose", "php"}, "", "", 0, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			target, dir, timeout, err := parseTestArgs(c.args)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if c.wantErr {
				return
			}
			if target != c.wantTarget || dir != c.wantDir || timeout != c.wantTimeout {
				t.Errorf("= (%q, %q, %s), want (%q, %q, %s)", target, dir, timeout, c.wantTarget, c.wantDir, c.wantTimeout)
			}
		})
	}
}

func TestRunCheck(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ok.Close()
	ko := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ko.Close()

	t.Run("toutes les cibles OK", func(t *testing.T) {
		var out bytes.Buffer
		if code := runCheck([]string{ok.URL}, &out); code != 0 {
			t.Errorf("code = %d, want 0 (sortie: %s)", code, out.String())
		}
		if !strings.HasPrefix(out.String(), "OK  "+ok.URL) {
			t.Errorf("sortie = %q, want une ligne \"OK  %s ...\"", out.String(), ok.URL)
		}
	})

	t.Run("une cible KO suffit à échouer, toutes sont vérifiées", func(t *testing.T) {
		var out bytes.Buffer
		if code := runCheck([]string{ko.URL, ok.URL}, &out); code != 1 {
			t.Errorf("code = %d, want 1", code)
		}
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		if len(lines) != 2 || !strings.HasPrefix(lines[0], "KO") || !strings.HasPrefix(lines[1], "OK") {
			t.Errorf("sortie = %q, want une ligne KO puis une ligne OK", out.String())
		}
	})

	t.Run("--status appliqué", func(t *testing.T) {
		var out bytes.Buffer
		if code := runCheck([]string{"--status", "500", ko.URL}, &out); code != 0 {
			t.Errorf("code = %d, want 0 avec --status 500 (sortie: %s)", code, out.String())
		}
	})

	t.Run("sans cible", func(t *testing.T) {
		var out bytes.Buffer
		if code := runCheck(nil, &out); code != 1 || !strings.Contains(out.String(), "usage") {
			t.Errorf("code = %d, sortie = %q, want 1 et l'usage", code, out.String())
		}
	})

	t.Run("option invalide", func(t *testing.T) {
		var out bytes.Buffer
		if code := runCheck([]string{"--status", "abc", ok.URL}, &out); code != 1 {
			t.Errorf("code = %d, want 1", code)
		}
	})
}

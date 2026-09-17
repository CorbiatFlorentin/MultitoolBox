package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

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

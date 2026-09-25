package runner

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestHttpDetect(t *testing.T) {
	var r httpRunner

	detected := []struct {
		name string
		file []string
	}{
		{".hurl à la racine", []string{"api.hurl"}},
		{".hurl dans un sous-dossier", []string{"tests", "api", "users.hurl"}},
		{"collection Postman", []string{"api.postman_collection.json"}},
	}
	for _, c := range detected {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			touch(t, dir, c.file...)
			if !r.Detect(dir) {
				t.Errorf("Detect() = false, want true avec %s", filepath.Join(c.file...))
			}
		})
	}

	notDetected := []struct {
		name string
		file []string
	}{
		{"dossier vide", nil},
		{"JSON quelconque", []string{"data.json"}},
		{".hurl dans node_modules", []string{"node_modules", "pkg", "x.hurl"}},
		{".hurl dans vendor", []string{"vendor", "x.hurl"}},
		{".hurl trop profond", []string{"a", "b", "c", "x.hurl"}},
	}
	for _, c := range notDetected {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if c.file != nil {
				touch(t, dir, c.file...)
			}
			if r.Detect(dir) {
				t.Errorf("Detect() = true, want false (%s)", c.name)
			}
		})
	}
}

func TestFindHTTPTestsSortedAndRelative(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "z.hurl")
	touch(t, dir, "tests", "a.hurl")
	touch(t, dir, "b.postman_collection.json")

	hurl, postman := findHTTPTests(dir)
	wantHurl := []string{filepath.Join("tests", "a.hurl"), "z.hurl"}
	if !reflect.DeepEqual(hurl, wantHurl) {
		t.Errorf("hurl = %v, want %v", hurl, wantHurl)
	}
	if want := []string{"b.postman_collection.json"}; !reflect.DeepEqual(postman, want) {
		t.Errorf("postman = %v, want %v", postman, want)
	}
}

func TestHttpTestInvokesHurl(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "health.hurl")
	touch(t, dir, "tests", "users.hurl")
	touch(t, dir, "api.postman_collection.json") // ignorée : hurl prioritaire
	hurlOut := withFakeCommand(t, "hurl", 0)
	newmanOut := withFakeCommand(t, "newman", 0)

	var r httpRunner
	if err := r.Test(dir); err != nil {
		t.Fatalf("Test() a retourné une erreur inattendue: %v", err)
	}
	want := "--test health.hurl " + filepath.Join("tests", "users.hurl")
	if got := readOut(t, hurlOut); got != want {
		t.Errorf("arguments passés à hurl = %q, want %q", got, want)
	}
	if _, err := os.Stat(newmanOut); err == nil {
		t.Error("newman a été appelé alors que des fichiers .hurl existent")
	}
}

func TestHttpTestInvokesNewman(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "api.postman_collection.json")
	out := withFakeCommand(t, "newman", 0)

	var r httpRunner
	if err := r.Test(dir); err != nil {
		t.Fatalf("Test() a retourné une erreur inattendue: %v", err)
	}
	if got, want := readOut(t, out), "run api.postman_collection.json"; got != want {
		t.Errorf("arguments passés à newman = %q, want %q", got, want)
	}
}

func TestHttpTestPropagatesFailure(t *testing.T) {
	t.Run("hurl", func(t *testing.T) {
		dir := t.TempDir()
		touch(t, dir, "api.hurl")
		withFakeCommand(t, "hurl", 1)
		var r httpRunner
		if err := r.Test(dir); err == nil {
			t.Error("Test() = nil, want une erreur quand hurl échoue")
		}
	})

	t.Run("newman mentionne la collection", func(t *testing.T) {
		dir := t.TempDir()
		touch(t, dir, "api.postman_collection.json")
		withFakeCommand(t, "newman", 1)
		var r httpRunner
		err := r.Test(dir)
		if err == nil || !strings.Contains(err.Error(), "api.postman_collection.json") {
			t.Errorf("Test() = %v, want une erreur nommant la collection en échec", err)
		}
	})
}

func TestHttpTestNothingToRun(t *testing.T) {
	var r httpRunner
	if err := r.Test(t.TempDir()); err == nil {
		t.Error("Test() = nil, want une erreur sans fichier .hurl ni collection")
	}
}

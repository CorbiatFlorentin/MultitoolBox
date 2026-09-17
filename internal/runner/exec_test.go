package runner

import "testing"

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "present.txt", "")

	if !fileExists(dir, "present.txt") {
		t.Error("fileExists() = false, want true pour un fichier présent")
	}
	if fileExists(dir, "absent.txt") {
		t.Error("fileExists() = true, want false pour un fichier absent")
	}
}

package runner

import (
	"strings"
	"testing"
	"time"
)

func TestRunTimeout(t *testing.T) {
	withSlowFakeCommand(t, "lent")
	SetTimeout(300 * time.Millisecond)
	t.Cleanup(func() { SetTimeout(0) })

	start := time.Now()
	err := run(t.TempDir(), "lent")
	if err == nil || !strings.Contains(err.Error(), "délai dépassé") {
		t.Fatalf("run() = %v, want une erreur \"délai dépassé\"", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("run() a mis %s à rendre la main, want ~300ms", elapsed)
	}
}

func TestRunWithoutTimeoutKeepsExitError(t *testing.T) {
	withFakeCommand(t, "rapide-ko", 3)
	SetTimeout(10 * time.Second)
	t.Cleanup(func() { SetTimeout(0) })

	err := run(t.TempDir(), "rapide-ko")
	if err == nil {
		t.Fatal("run() = nil, want l'erreur de sortie de la commande")
	}
	if strings.Contains(err.Error(), "délai dépassé") {
		t.Errorf("run() = %v, un échec rapide ne doit pas être signalé comme un délai dépassé", err)
	}
}

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

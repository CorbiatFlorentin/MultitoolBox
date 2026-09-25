package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// withFakeCommand installe sur le PATH un faux exécutable nommé `name` qui
// écrit ses arguments dans un fichier puis se termine avec exitCode. Elle
// renvoie le chemin du fichier de sortie, à lire avec readOut.
func withFakeCommand(t *testing.T, name string, exitCode int) string {
	t.Helper()
	binDir := t.TempDir()
	outFile := filepath.Join(binDir, name+".out")

	var scriptPath, content string
	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(binDir, name+".cmd")
		content = fmt.Sprintf("@echo off\r\necho %%*>\"%s\"\r\nexit /b %d\r\n", outFile, exitCode)
	} else {
		scriptPath = filepath.Join(binDir, name)
		content = fmt.Sprintf("#!/bin/sh\necho \"$@\" > \"%s\"\nexit %d\n", outFile, exitCode)
	}
	if err := os.WriteFile(scriptPath, []byte(content), 0o755); err != nil {
		t.Fatalf("écriture du faux exécutable %s: %v", name, err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return outFile
}

// withSlowFakeCommand installe sur le PATH un faux exécutable nommé `name`
// qui dort une dizaine de secondes, pour tester le délai maximal de run.
func withSlowFakeCommand(t *testing.T, name string) {
	t.Helper()
	binDir := t.TempDir()

	var scriptPath, content string
	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(binDir, name+".cmd")
		content = "@echo off\r\nping -n 11 127.0.0.1 >nul\r\n"
	} else {
		scriptPath = filepath.Join(binDir, name)
		content = "#!/bin/sh\nsleep 10\n"
	}
	if err := os.WriteFile(scriptPath, []byte(content), 0o755); err != nil {
		t.Fatalf("écriture du faux exécutable %s: %v", name, err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// readOut lit et renvoie le contenu (sans espaces superflus) écrit par un
// faux exécutable créé via withFakeCommand.
func readOut(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("lecture de %s: %v", path, err)
	}
	return strings.TrimSpace(string(data))
}

package runner

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

func init() {
	Register(httpRunner{})
}

// httpRunner lance des tests d'API décrits en fichiers : .hurl (via hurl)
// ou collections Postman (via newman). Hurl est prioritaire si les deux
// sont présents.
type httpRunner struct{}

func (httpRunner) Name() string { return "http" }

func (httpRunner) Detect(dir string) bool {
	hurl, postman := findHTTPTests(dir)
	return len(hurl) > 0 || len(postman) > 0
}

func (httpRunner) Test(dir string) error {
	hurl, postman := findHTTPTests(dir)
	if len(hurl) > 0 {
		return run(dir, "hurl", append([]string{"--test"}, hurl...)...)
	}
	if len(postman) == 0 {
		return fmt.Errorf("aucun fichier .hurl ni collection Postman trouvé dans %s", dir)
	}
	for _, c := range postman {
		if err := run(dir, "newman", "run", c); err != nil {
			return fmt.Errorf("%s: %w", c, err)
		}
	}
	return nil
}

// httpMaxDepth limite la recherche (dir = 0) pour ne pas parcourir tout un
// monorepo : les tests d'API sont en général à la racine ou dans un dossier
// dédié (tests/, api/...).
const httpMaxDepth = 2

var httpSkipDirs = map[string]bool{"node_modules": true, "vendor": true, ".git": true}

// findHTTPTests renvoie, triés et relatifs à dir, les fichiers .hurl et les
// collections Postman (*.postman_collection.json).
func findHTTPTests(dir string) (hurl, postman []string) {
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		if d.IsDir() {
			if rel != "." && (httpSkipDirs[d.Name()] || strings.Count(rel, string(filepath.Separator)) >= httpMaxDepth) {
				return filepath.SkipDir
			}
			return nil
		}
		switch {
		case strings.HasSuffix(d.Name(), ".hurl"):
			hurl = append(hurl, rel)
		case strings.HasSuffix(d.Name(), ".postman_collection.json"):
			postman = append(postman, rel)
		}
		return nil
	})
	sort.Strings(hurl)
	sort.Strings(postman)
	return hurl, postman
}

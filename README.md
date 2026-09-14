# multitest

CLI unique pour lancer les tests d'un projet, quel que soit son langage.
Détecte automatiquement le type de projet dans un dossier et délègue au vrai
outil de test (pytest, phpunit, tsc/vitest, npm test...).

Langages supportés aujourd'hui : **Python, React, PHP, TypeScript**.

## Usage

<!-- usage:start -->
```
multitest - lance les tests d'un projet, quel que soit son langage

Usage:
  multitest                               (mode interactif)
  multitest test <langage|all> [chemin]   (langages: python, react, php, typescript)
  multitest list                          (langages supportés)
  multitest help                          (affiche cette aide)
```
<!-- usage:end -->

Sans `chemin`, le dossier courant est utilisé. Sans aucun argument, l'outil
s'ouvre en mode interactif (menu au clavier).

## Build

```bash
go build -o multitest.exe .
```

## Tests

```bash
go test ./... -v
```

## CI

Chaque push/PR sur `main` déclenche `.github/workflows/ci.yml` :
- build, `go vet`, `go test`, vérification du formatage (`gofmt`)
- compilation croisée (windows-amd64, linux-amd64, darwin-amd64, darwin-arm64),
  publiée en artefacts téléchargeables sur le run GitHub Actions
- la section "Usage" ci-dessus est régénérée automatiquement à partir de
  `multitest help` et commitée si elle a changé

## Ajouter un langage

Implémenter l'interface `runner.Runner` (`Name`, `Detect`, `Test`) dans
`internal/runner/<langage>.go` et l'enregistrer dans son `init()` via
`runner.Register(...)`. Voir `internal/runner/python.go` pour un exemple minimal.

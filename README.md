# multitest

CLI unique pour lancer les tests d'un projet, quel que soit son langage.
Détecte automatiquement le type de projet dans un dossier et délègue au vrai
outil de test (pytest, phpunit, tsc/vitest, npm test...).

## Usage

```bash
multitest list                     # langages supportés
multitest test all [chemin]        # détecte et lance tout ce qui matche
multitest test python [chemin]
multitest test react [chemin]
multitest test php [chemin]
multitest test typescript [chemin]
```

Sans `chemin`, le dossier courant est utilisé.

## Build

```bash
go build -o multitest.exe .
```

## Ajouter un langage

Implémenter l'interface `runner.Runner` (`Name`, `Detect`, `Test`) dans
`internal/runner/<langage>.go` et l'enregistrer dans son `init()` via
`runner.Register(...)`. Voir `internal/runner/python.go` pour un exemple minimal.

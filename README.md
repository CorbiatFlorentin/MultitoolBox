# multitest

[![CI](https://github.com/CorbiatFlorentin/MultitoolBox/actions/workflows/ci.yml/badge.svg)](https://github.com/CorbiatFlorentin/MultitoolBox/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

CLI unique pour lancer les tests d'un projet, quel que soit son langage.
Détecte automatiquement le type de projet dans un dossier et délègue au vrai
outil de test (pytest, phpunit, tsc, npm test...) plutôt que de le
réimplémenter.

Langages supportés aujourd'hui : **Python, React, PHP, TypeScript**, plus des
tests d'API **HTTP** (fichiers `.hurl` ou collections Postman) et un smoke test
réseau intégré (`multitest check`). Le tout
est pensé pour être facilement extensible à d'autres langages (voir
[Ajouter un langage](#ajouter-un-langage)).

## Pourquoi ce projet

Petit outil perso pour éviter de jongler entre `pytest`, `phpunit`, `npm test`
et `tsc` selon le projet sur lequel je travaille, et pour pratiquer :
- Go (architecture à plugins, cross-compilation)
- un pipeline CI/CD complet (tests, artefacts multi-plateformes, release
  automatique, doc auto-générée)

## Démo

```
$ multitest list
php
python
react
typescript
http

$ multitest test all examples/demo-ts
== typescript ==
2 passed, 0 failed
```

Le dossier [`examples/demo-ts`](examples/demo-ts) est un mini-projet
TypeScript fonctionnel que tu peux tester toi-même :

```bash
cd examples/demo-ts && npm install
cd ../.. && ./multitest test all examples/demo-ts
```

## Usage

<!-- usage:start -->
```
multitest - lance les tests d'un projet, quel que soit son langage

Usage:
  multitest                               (mode interactif)
  multitest test [--timeout 5m] <langage|all> [chemin]
                                          (langages: python, react, php, typescript, http)
  multitest check [--timeout 5s] [--status 200] <url|host:port>...
                                          (smoke test réseau : HTTP ou port TCP)
  multitest list                          (langages supportés)
  multitest help                          (affiche cette aide)
```
<!-- usage:end -->

Sans `chemin`, le dossier courant est utilisé. Sans aucun argument, l'outil
s'ouvre en mode interactif (menu au clavier). Les options (`--timeout`,
`--status`) se placent avant les arguments.

### PHP

Le lanceur est choisi dans cet ordre : `php artisan test` (Laravel),
`vendor/bin/pest`, `vendor/bin/phpunit` (tous lancés via `php`, ce qui marche
aussi sous Windows), puis le script `test` de `composer.json`. Un
`composer.json` sans aucun signe de tests (pas de `tests/`, de phpunit/pest ni
de script `test`) n'est pas détecté par `test all`. Mini-projet d'exemple :
[`examples/demo-php`](examples/demo-php) (`composer install` puis
`multitest test php examples/demo-php`).

### Réseau

- `multitest test http` lance `hurl --test` sur les fichiers `.hurl` trouvés
  (jusqu'à 2 niveaux de sous-dossiers, hors `node_modules`/`vendor`), sinon
  `newman run` sur chaque `*.postman_collection.json`.
- `multitest check` fait un GET sur une URL (OK si code < 400, ou égal à
  `--status` ; les redirections ne sont pas suivies) ou ouvre une connexion
  TCP sur `host:port`. Code de sortie 1 si une cible est KO.

```
$ multitest check https://github.com github.com:443 localhost:1
OK  https://github.com  HTTP 200 (147ms)
OK  github.com:443  port ouvert (20ms)
KO  localhost:1  dial tcp [::1]:1: connectex: No connection could be made ... (2ms)
```

## Installation

Binaires précompilés (windows/linux/macOS) disponibles sur la page
[Releases](../../releases). Sinon, depuis les sources :

```bash
go build -o multitest.exe .
```

## Tests

```bash
go test ./... -v
```

## CI

Chaque push/PR sur `main` déclenche `.github/workflows/ci.yml` :
- build, `go vet`, `go test` (avec PHP + composer installés pour lancer le
  test d'intégration de `examples/demo-php`), vérification du formatage
  (`gofmt`)
- compilation croisée (windows-amd64, linux-amd64, darwin-amd64, darwin-arm64),
  publiée en artefacts téléchargeables sur le run GitHub Actions
- la section "Usage" ci-dessus est régénérée automatiquement à partir de
  `multitest help` et commitée si elle a changé

## Release

Pousser un tag `v*` (ex. `v0.1.0`) déclenche `.github/workflows/release.yml` :
il rebuild les 4 binaires (windows-amd64, linux-amd64, darwin-amd64,
darwin-arm64) et crée une [Release GitHub](../../releases) avec ces binaires
attachés et des notes générées automatiquement à partir des commits.

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Ajouter un langage

Implémenter l'interface `runner.Runner` (`Name`, `Detect`, `Test`) dans
`internal/runner/<langage>.go` et l'enregistrer dans son `init()` via
`runner.Register(...)`. Voir `internal/runner/python.go` pour un exemple minimal.

## Licence

[MIT](LICENSE)

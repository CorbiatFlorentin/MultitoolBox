package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"multitest/internal/check"
	"multitest/internal/runner"
)

func usage() {
	fmt.Println("multitest - lance les tests d'un projet, quel que soit son langage")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  multitest                               (mode interactif)")
	fmt.Println("  multitest test [--timeout 5m] <langage|all> [chemin]")
	fmt.Println("                                          (langages: python, react, php, typescript, http)")
	fmt.Println("  multitest check [--timeout 5s] [--status 200] <url|host:port>...")
	fmt.Println("                                          (smoke test réseau : HTTP ou port TCP)")
	fmt.Println("  multitest list                          (langages supportés)")
	fmt.Println("  multitest help                          (affiche cette aide)")
}

func main() {
	if len(os.Args) < 2 {
		interactive()
		return
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		usage()

	case "list":
		for name := range runner.All() {
			fmt.Println(name)
		}

	case "test":
		target, dir, timeout, err := parseTestArgs(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			usage()
			os.Exit(1)
		}
		runner.SetTimeout(timeout)
		if err := runTest(target, dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case "check":
		os.Exit(runCheck(os.Args[2:], os.Stdout))

	default:
		usage()
		os.Exit(1)
	}
}

// parseTestArgs lit "[--timeout d] <langage|all> [chemin]".
func parseTestArgs(args []string) (target, dir string, timeout time.Duration, err error) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.DurationVar(&timeout, "timeout", 0, "durée max de chaque commande de test (0 = illimité)")
	if err := fs.Parse(args); err != nil {
		return "", "", 0, err
	}
	switch fs.NArg() {
	case 1:
		return fs.Arg(0), ".", timeout, nil
	case 2:
		return fs.Arg(0), fs.Arg(1), timeout, nil
	}
	return "", "", 0, fmt.Errorf("usage: multitest test [--timeout 5m] <langage|all> [chemin]")
}

// runCheck vérifie chaque cible réseau, affiche une ligne OK/KO par cible et
// renvoie le code de sortie : 0 si tout est OK, 1 sinon (ou en cas d'erreur
// d'arguments).
func runCheck(args []string, w io.Writer) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(w)
	var opts check.Options
	fs.DurationVar(&opts.Timeout, "timeout", 5*time.Second, "délai max par cible")
	fs.IntVar(&opts.WantStatus, "status", 0, "code HTTP attendu (0 = tout code < 400)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() == 0 {
		fmt.Fprintln(w, "usage: multitest check [--timeout 5s] [--status 200] <url|host:port>...")
		return 1
	}

	code := 0
	for _, target := range fs.Args() {
		res := check.Target(target, opts)
		status := "OK"
		if !res.OK {
			status = "KO"
			code = 1
		}
		fmt.Fprintf(w, "%s  %s  %s (%s)\n", status, res.Target, res.Detail, res.Duration.Round(time.Millisecond))
	}
	return code
}

// runTest lance les tests pour target ("all" ou un nom de langage précis) dans dir.
func runTest(target, dir string) error {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return fmt.Errorf("dossier introuvable: %s", dir)
	}
	if target == "all" {
		ran := false
		for name, r := range runner.All() {
			if !r.Detect(dir) {
				continue
			}
			ran = true
			fmt.Printf("== %s ==\n", name)
			if err := r.Test(dir); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
		if !ran {
			fmt.Println("Aucun projet détecté dans", dir)
		}
		return nil
	}

	r, ok := runner.Get(target)
	if !ok {
		return fmt.Errorf("langage inconnu: %s", target)
	}
	return r.Test(dir)
}

func interactive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== multitest ===")
	fmt.Println("Lance les tests d'un projet, quel que soit son langage (python, react, php, typescript, http, all).")
	fmt.Println("Tape 'list' pour voir les langages supportés, 'q' pour quitter.")
	fmt.Println()

	for {
		fmt.Print("Langage : ")
		lang := readLine(reader)

		switch lang {
		case "":
			continue

		case "q", "quit", "exit":
			fmt.Println("À bientôt !")
			return

		case "list":
			for name := range runner.All() {
				fmt.Println(" -", name)
			}

		default:
			fmt.Print("Chemin du projet (vide = dossier courant) : ")
			dir := readLine(reader)
			if dir == "" {
				dir = "."
			}

			if err := runTest(lang, dir); err != nil {
				fmt.Println("Erreur:", err)
			}
		}

		fmt.Println()
	}
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

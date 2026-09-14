package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"multitest/internal/runner"
)

func usage() {
	fmt.Println("multitest - lance les tests d'un projet, quel que soit son langage")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  multitest                               (mode interactif)")
	fmt.Println("  multitest test <langage|all> [chemin]   (langages: python, react, php, typescript)")
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
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		target := os.Args[2]
		dir := "."
		if len(os.Args) >= 4 {
			dir = os.Args[3]
		}
		if err := runTest(target, dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	default:
		usage()
		os.Exit(1)
	}
}

// runTest lance les tests pour target ("all" ou un nom de langage précis) dans dir.
func runTest(target, dir string) error {
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
	fmt.Println("Lance les tests d'un projet, quel que soit son langage (python, react, php, typescript).")
	fmt.Println()

	for {
		fmt.Println("1) Lister les langages supportés")
		fmt.Println("2) Tester un projet")
		fmt.Println("3) Quitter")
		fmt.Print("> ")

		switch readLine(reader) {
		case "1":
			for name := range runner.All() {
				fmt.Println(" -", name)
			}

		case "2":
			fmt.Print("Langage (python/react/php/typescript/all) : ")
			lang := readLine(reader)

			fmt.Print("Chemin du projet (vide = dossier courant) : ")
			dir := readLine(reader)
			if dir == "" {
				dir = "."
			}

			if err := runTest(lang, dir); err != nil {
				fmt.Println("Erreur:", err)
			}

		case "3", "q", "quit", "exit":
			fmt.Println("À bientôt !")
			return

		default:
			fmt.Println("Choix invalide.")
		}

		fmt.Println()
	}
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

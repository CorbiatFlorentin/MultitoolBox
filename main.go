package main

import (
	"fmt"
	"os"

	"multitest/internal/runner"
)

func usage() {
	fmt.Println("multitest - lance les tests d'un projet, quel que soit son langage")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  multitest test <langage|all> [chemin]   (langages: python, react, php, typescript)")
	fmt.Println("  multitest list                          (langages supportés)")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
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

		if target == "all" {
			ran := false
			for name, r := range runner.All() {
				if r.Detect(dir) {
					ran = true
					fmt.Printf("== %s ==\n", name)
					if err := r.Test(dir); err != nil {
						fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
						os.Exit(1)
					}
				}
			}
			if !ran {
				fmt.Println("Aucun projet détecté dans", dir)
			}
			return
		}

		r, ok := runner.Get(target)
		if !ok {
			fmt.Fprintf(os.Stderr, "langage inconnu: %s\n", target)
			usage()
			os.Exit(1)
		}
		if err := r.Test(dir); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

	default:
		usage()
		os.Exit(1)
	}
}

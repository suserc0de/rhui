package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/suserc0de/rhui/parser"
)

func printUsage() {
	fmt.Println("RHUI CLI Tooling (v1.1.0)")
	fmt.Println("\nUsage:")
	fmt.Println("  rhui <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  fmt    Format and indent .rhui template files")
	fmt.Println("  check  Validate syntax and lint .rhui templates without building")
	fmt.Println("  help   Show this help message")
	fmt.Println("\nRun 'rhui <command> -h' for more information on a command.")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "fmt":
		runFmt(os.Args[2:])
	case "check", "lint":
		runCheck(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runFmt(args []string) {
	fsSet := flag.NewFlagSet("fmt", flag.ExitOnError)
	writeInPlace := fsSet.Bool("w", false, "Write result to (source) file instead of stdout")
	dryRun := fsSet.Bool("d", false, "Display diffs instead of rewriting files")
	fsSet.Usage = func() {
		fmt.Println("Usage: rhui fmt [flags] [path ...]")
		fmt.Println("\nFlags:")
		fsSet.PrintDefaults()
	}
	_ = fsSet.Parse(args)

	targets := fsSet.Args()
	if len(targets) == 0 {
		targets = []string{"."}
	}

	files, err := collectRhuiFiles(targets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting files: %v\n", err)
		os.Exit(1)
	}

	formatter := parser.NewFormatter()
	hasChanges := false

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			continue
		}

		formatted, err := formatter.Format(file, string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Syntax error in %s: %v\n", file, err)
			os.Exit(1)
		}

		if string(content) != formatted {
			hasChanges = true
			if *dryRun {
				fmt.Printf("File needs formatting: %s\n", file)
			} else if *writeInPlace {
				if err := os.WriteFile(file, []byte(formatted), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", file, err)
				} else {
					fmt.Printf("Formatted: %s\n", file)
				}
			} else {
				fmt.Print(formatted)
			}
		}
	}

	if *dryRun && hasChanges {
		os.Exit(1)
	}
}

func runCheck(args []string) {
	fsSet := flag.NewFlagSet("check", flag.ExitOnError)
	fsSet.Usage = func() {
		fmt.Println("Usage: rhui check [path ...]")
		fmt.Println("\nChecks and validates syntax across all .rhui template files.")
	}
	_ = fsSet.Parse(args)

	targets := fsSet.Args()
	if len(targets) == 0 {
		targets = []string{"."}
	}

	files, err := collectRhuiFiles(targets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting files: %v\n", err)
		os.Exit(1)
	}

	totalErrors := 0
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			totalErrors++
			continue
		}

		p := parser.NewParser(parser.NewLexer(file, string(content)))
		_, err = p.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
			totalErrors++
		}
		for _, parseErr := range p.Errors() {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, parseErr)
			totalErrors++
		}
	}

	if totalErrors > 0 {
		fmt.Fprintf(os.Stderr, "\nValidation failed: found %d error(s) in %d file(s)\n", totalErrors, len(files))
		os.Exit(1)
	}

	fmt.Printf("Validation passed: %d .rhui file(s) checked successfully with zero errors.\n", len(files))
}

func collectRhuiFiles(targets []string) ([]string, error) {
	var files []string
	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if strings.HasSuffix(target, ".rhui") {
				files = append(files, target)
			}
			continue
		}

		err = filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// Skip hidden dirs (.git, .vscode, etc.)
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".rhui") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

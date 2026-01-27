package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"leakcheck/internal/analyzer"
	"leakcheck/internal/parser"
	"leakcheck/internal/reporter"
	"leakcheck/internal/scanner"
)

var (
	version = "2.0.0"
)

func main() {
	// Define flags
	excludeFlag := flag.String("exclude", "", "Comma-separated list of directories to exclude (e.g., vendor,build,third_party)")
	jsonFlag := flag.Bool("json", false, "Output results in JSON format")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	helpFlag := flag.Bool("help", false, "Show help message")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: leakcheck [options] <path> [paths...]\n\n")
		fmt.Fprintf(os.Stderr, "C++ Memory Leak Detector - Static analysis tool to detect potential memory leaks\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  leakcheck ./src                    Scan all C++ files in ./src\n")
		fmt.Fprintf(os.Stderr, "  leakcheck --exclude=vendor ./      Scan all files, excluding vendor directory\n")
		fmt.Fprintf(os.Stderr, "  leakcheck --json ./src > out.json  Output results as JSON\n")
	}

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Printf("leakcheck version %s\n", version)
		os.Exit(0)
	}

	// Get paths to scan
	paths := flag.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No paths specified")
		fmt.Fprintln(os.Stderr, "Run 'leakcheck --help' for usage")
		os.Exit(1)
	}

	// Parse exclude patterns
	var excludes []string
	if *excludeFlag != "" {
		excludes = strings.Split(*excludeFlag, ",")
		for i := range excludes {
			excludes[i] = strings.TrimSpace(excludes[i])
		}
	}

	// Scan for C++ files
	s := scanner.NewScanner(excludes)
	files, err := s.ScanPaths(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning paths: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No C++ files found")
		os.Exit(0)
	}

	if !*jsonFlag {
		fmt.Printf("Scanning %d file(s)...\n", len(files))
	}

	// Parse all files and register classes
	registry := parser.NewClassRegistry()
	for _, file := range files {
		classes, err := parser.ParseFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error parsing %s: %v\n", file, err)
			continue
		}
		registry.AddClasses(classes)
	}

	// Merge classes from headers and implementations
	allClasses := registry.MergeClasses()

	if !*jsonFlag {
		fmt.Printf("Found %d class(es) with pointer members\n", countClassesWithPointers(allClasses))
	}

	// Analyze for leaks
	leaks := analyzer.AnalyzeClasses(allClasses)

	// Report results
	r := reporter.NewReporter(os.Stdout, *jsonFlag)
	if err := r.Report(leaks); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
		os.Exit(1)
	}

	// Exit with error code if leaks found
	if len(leaks) > 0 {
		os.Exit(1)
	}
}

func countClassesWithPointers(classes []parser.Class) int {
	count := 0
	for _, c := range classes {
		for _, m := range c.Members {
			if m.IsPointer {
				count++
				break
			}
		}
	}
	return count
}

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/subhasundardas/gofar/cmd/gofar/gen"
)

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 0
	}

	// Check for required tools
	if err := checkTool("air", "go install github.com/air-verse/air@latest"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := checkTool("templ", "go install github.com/a-h/templ/cmd/templ@latest"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	command := args[0]
	flags, err := parseArgs(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	switch command {
	case "make:module":
		err = gen.MakeModule(flags)
	case "make:handler":
		err = gen.MakeHandler(flags)
	case "make:service":
		err = gen.MakeService(flags)
	case "make:repository":
		err = gen.MakeRepository(flags)
	case "list:modules":
		err = gen.ListModules()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", command)
		printUsage()
		return 1
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func parseArgs(raw []string) (map[string]string, error) {
	out := make(map[string]string)
	for _, arg := range raw {
		found := false
		for i, c := range arg {
			if c == '=' {
				out[arg[:i]] = arg[i+1:]
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid argument %q — expected key=value", arg)
		}
	}
	return out, nil
}

func printUsage() {
	fmt.Print(`GoFar CLI — module and scaffold generator

Usage:
  go run ./cmd/gofar <command> [key=value ...]

Commands:
  make:module  name=<Name>               Generate a complete module scaffold
  make:handler name=<Name> module=<Mod>  Add a handler to an existing module
  make:service name=<Name> module=<Mod>  Add a service to an existing module
  make:repository  name=<Name> module=<Mod>  ← add repo to existing module
  list:modules                           List all modules in modules/all/all.go

Examples:
  go run ./cmd/gofar make:module  name=Billing
  go run ./cmd/gofar make:module  name=Notifications
  go run ./cmd/gofar make:handler name=Refund module=Billing
  go run ./cmd/gofar make:service name=TaxCalculator module=Billing

You can also add a Makefile target:
  make module name=Billing   →  go run ./cmd/gofar make:module name=Billing
`)
}

// Helper functions

func checkTool(name, installCmd string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s not found. Install with: %s", name, installCmd)
	}
	return nil
}

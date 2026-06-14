// Package main is the GoFar CLI tool.
//
// Usage:
//
//	go run ./cmd/gofar <command> [flags]
//
// Commands:
//
//	make:module     name=<Name>              Generate a full module scaffold
//	make:handler    name=<Name> module=<Mod> Add a handler to an existing module
//	make:service    name=<Name> module=<Mod> Add a service to an existing module
//	make:repository name=<Name> module=<Mod> Add a repository to an existing module
//	list:modules                           Print all registered modules
//
// Examples:
//
//	go run ./cmd/gofar make:module     name=Billing
//	go run ./cmd/gofar make:module     name=Notifications
//	go run ./cmd/gofar make:handler    name=Invoice module=Billing
//	go run ./cmd/gofar make:service    name=TaxCalculator module=Billing
//	go run ./cmd/gofar make:repository name=Invoice module=Billing
package main

import (
	"fmt"
	"os"

	"github.com/subhasundardas/gofar/cmd/gofar/gen"
)

func main() {
	os.Exit(run(os.Args[1:]))

}

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

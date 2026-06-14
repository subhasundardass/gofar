package main

import (
	"fmt"
	"os/exec"
)

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

// Package main is the GoFar CLI tool.
//
// Usage:
//
//	go run ./cmd/gofar <command> [flags]
//
// Commands:
//
//	make:module  name=<Name>              Generate a full module scaffold
//	make:handler name=<Name> module=<Mod> Add a handler to an existing module
//	make:service name=<Name> module=<Mod> Add a service to an existing module
//	list:modules                          Print all registered modules
//
// Examples:
//
//	go run ./cmd/gofar make:module name=Billing
//	go run ./cmd/gofar make:module name=Notifications
//	go run ./cmd/gofar make:handler name=Invoice module=Billing
package main

import (
	"os"
)

func main() {
	os.Exit(run(os.Args[1:]))

}

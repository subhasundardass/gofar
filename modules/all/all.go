// Package all registers every application module in boot order.
// Import this package with a blank import from cmd/server/main.go:
//
//	import _ "github.com/subhasundardas/gofar/modules/all"
//
// To add a module: add one import line below.
// To disable a module: set "enabled": false in its manifest.json.
package all

import (
	_ "github.com/subhasundardas/gofar/modules/accounting"
	_ "github.com/subhasundardas/gofar/modules/base"
	_ "github.com/subhasundardas/gofar/modules/example"
	// modules registered below — order = boot order
)

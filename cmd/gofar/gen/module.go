package gen

import (
	"fmt"
	"os"
	"strings"
)

// MakeModule is the handler for `go run ./cmd/gofar make:module name=<Name>`.
// It generates the full module scaffold under modules/<snake_name>/ and
// registers the module in modules/all/all.go.
//
// Files created:
//
//	modules/<snake>/manifest.json
//	modules/<snake>/module.go
//	modules/<snake>/routes.go
//	modules/<snake>/events.go
//	modules/<snake>/repository/repository.go
//	modules/<snake>/service/service.go
//	modules/<snake>/handler/handler.go
//
// Then appends one import line to modules/all/all.go.
func MakeModule(args map[string]string) error {
	name, ok := args["name"]
	if !ok || name == "" {
		return fmt.Errorf("name is required: make:module name=<Name>")
	}

	mod := newNames(name)
	modPath, err := readModulePath()
	if err != nil {
		return err
	}

	data := templateData{Module: mod, ModPath: modPath}
	base := fmt.Sprintf("modules/%s", mod.Snake)

	fmt.Printf("\n🔧 Generating module \"%s\"\n\n", mod.Pascal)

	files := []struct {
		path string
		tmpl string
	}{
		{base + "/manifest.json", tmplManifest},
		{base + "/module.go", tmplModule},
		{base + "/routes.go", tmplRoutes},
		{base + "/events.go", tmplEvents},
		{base + "/repository/repository.go", tmplRepository},
		{base + "/service/service.go", tmplService},
		{base + "/handler/handler.go", tmplHandler},
	}

	for _, f := range files {
		content, err := renderTemplate(f.path, f.tmpl, data)
		if err != nil {
			return err
		}
		if err := writeFile(f.path, content); err != nil {
			return err
		}
	}

	// Register in modules/all/all.go
	if err := registerInAll(mod, modPath); err != nil {
		return err
	}

	fmt.Printf("\n✅ Module \"%s\" generated successfully!\n", mod.Pascal)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Add your Ent schema in ent/schema/%s.go\n", mod.Snake)
	fmt.Printf("  2. Run: go generate ./ent/...\n")
	fmt.Printf("  3. Wire the *ent.Client in modules/%s/module.go Register()\n", mod.Snake)
	fmt.Printf("  4. Run: go build ./...\n\n")

	return nil
}

// registerInAll appends a blank import line to modules/all/all.go.
// Creates the file if it doesn't exist yet.
func registerInAll(mod Names, modPath string) error {
	allPath := "modules/all/all.go"

	// Create the file if it does not exist
	if _, err := os.Stat(allPath); os.IsNotExist(err) {
		if err := os.MkdirAll("modules/all", 0755); err != nil {
			return err
		}
		skeleton := fmt.Sprintf(`// Package all registers every application module in boot order.
// Import this package with a blank import from cmd/server/main.go:
//
//	import _ "%s/modules/all"
//
// To add a module: add one import line below.
// To disable a module: set "enabled": false in its manifest.json.
package all

import (
	// modules registered below — order = boot order
)
`, modPath)
		if err := os.WriteFile(allPath, []byte(skeleton), 0644); err != nil {
			return fmt.Errorf("create %s: %w", allPath, err)
		}
		fmt.Printf("  ✓ created  %s\n", allPath)
	}

	importLine := fmt.Sprintf(`	_ "%s/modules/%s"`, modPath, mod.Snake)
	return appendToFile(allPath, "// modules registered below — order = boot order", importLine)
}

// readModulePath reads the module path from go.mod in the current directory.
// e.g. "github.com/subhasundardas/gofar"
func readModulePath() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("could not open go.mod — run this command from the project root: %w", err)
	}

	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}
	return "", fmt.Errorf("could not find module path in go.mod")
}

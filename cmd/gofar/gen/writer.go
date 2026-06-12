package gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// writeFile writes content to path, creating parent directories as needed.
// If the file already exists it is NOT overwritten — an error is returned
// instead so the user never loses hand-edited code accidentally.
func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s (delete it first to regenerate)", path)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	fmt.Printf("  ✓ created  %s\n", path)
	return nil
}

// renderTemplate executes a text/template string with data and returns the result.
func renderTemplate(name, tmpl string, data any) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}
	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", name, err)
	}
	return sb.String(), nil
}

// appendToFile appends line to the file at path if line is not already present.
// Used to add the blank import to modules/all/all.go without duplicating it.
func appendToFile(path, marker, line string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	content := string(data)
	if strings.Contains(content, line) {
		fmt.Printf("  ~ skipped  %s (already registered)\n", path)
		return nil
	}
	// Insert before the closing ) of the import block that contains marker
	updated := strings.Replace(content, marker, line+"\n"+marker, 1)
	if updated == content {
		// marker not found — just append the line at the end
		updated = content + "\n" + line
	}
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	fmt.Printf("  ✓ updated  %s\n", path)
	return nil
}

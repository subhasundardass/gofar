//go:build !unix && !linux && !darwin && !windows

package logger

// isatty on unsupported platforms always returns false, which means colors
// are disabled by default. Callers that really want color on these
// platforms can still use SetColor(true) explicitly.
func isatty(fd uintptr) bool { return false }

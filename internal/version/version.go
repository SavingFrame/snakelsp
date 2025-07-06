// Package version provides version information for the SnakeLSP server.
// It contains the current version string and utilities to retrieve it.
package version

var Version = "dev"

func Get() string {
	return Version
}

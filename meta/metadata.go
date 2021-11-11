// Package meta contains metadata about the program itself
package meta

var (
	ConfigFile            string  // Path to the config file, if any
	ProgramDir            string  // Directory that contains the running program
	ProgramFilename       string  // Base filename of the running program
	ModuleName            string  // Name of the Go module of this project, passed in by ldflags during build.
	Version               = "TBD" // Program version, populated by ldflags during build
	DefaultConfigFilename = "cloudflare-ddns.toml"
)

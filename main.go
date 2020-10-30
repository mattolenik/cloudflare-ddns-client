package main

import (
	"flag"
	"fmt"
	"os"
)

// version is populated by the ldflags argument during build.
var version string

func main() {
	// Setting arg 0 makes sure that -help output has the correct program name when being invoked with "go run"
	os.Args[0] = "cloudflare-ddns"
	var flagVersion bool
	flag.BoolVar(&flagVersion, "version", false, "Print the program version")

	flag.Parse()

	if flagVersion {
		PrintVersion()
	}
}

// PrintVersion prints the program version and exits.
func PrintVersion() {
	fmt.Println(version)
	os.Exit(0)
}

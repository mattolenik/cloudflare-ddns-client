package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/ip"
	"github.com/rs/zerolog/log"
)

// version is populated by the ldflags argument during build.
var version string

func main() {
	err := mainE()
	if err != nil {
		// use stack trace
		log.Error().Msg(err.Error())
		os.Exit(1)
	}
}

func mainE() error {
	// Setting arg 0 makes sure that -help output has the correct program name when being invoked with "go run"
	os.Args[0] = "cloudflare-ddns"
	var flagVersion bool
	flag.BoolVar(&flagVersion, "version", false, "Print the program version")

	flag.Parse()

	if flagVersion {
		PrintVersion()
	}

	ip, err := ip.GetExternalIP()
	if err != nil {
		return errors.Annotate(err, "unable to retrieve external IP")
	}
	fmt.Println(ip)
	return nil
}

// PrintVersion prints the program version and exits.
func PrintVersion() {
	fmt.Println(version)
	os.Exit(0)
}

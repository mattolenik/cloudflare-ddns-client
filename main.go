package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/ip"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
)

// version is populated by the ldflags argument during build.
var version string

const configFileName = "cloudflare-ddns.conf"

// Config represents TOML program configuration
type Config struct {
	Domain string `toml:"domain"`
	Record string `toml:"record"`
}

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
	var flagConfigPath string
	flag.BoolVar(&flagVersion, "version", false, "Print the program version")
	flag.StringVar(&flagConfigPath, "config", "/etc/"+configFileName, "Path to configuration file")

	flag.Parse()

	if flagVersion {
		printVersion()
	}
	config, err := loadConfig(flagConfigPath)
	if err != nil {
		return errors.Trace(err)
	}
	fmt.Println(*config)

	ip, err := ip.GetExternalIP()
	if err != nil {
		return errors.Annotate(err, "unable to retrieve external IP")
	}
	fmt.Println(ip)
	return nil
}

// loadConfig loads the TOML configuration at the specified path.
func loadConfig(path string) (*Config, error) {
	config := &Config{}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to read config file '%s", path)
	}
	err = toml.Unmarshal(contents, config)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to parse config file '%s", path)
	}
	return config, nil
}

// printVersion prints the program version and exits.
func printVersion() {
	fmt.Println(version)
	os.Exit(0)
}

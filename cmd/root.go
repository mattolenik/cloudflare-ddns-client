package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/conf"
	"github.com/mattolenik/cloudflare-ddns-client/ddns"
	"github.com/mattolenik/cloudflare-ddns-client/errhandler"
	"github.com/mattolenik/cloudflare-ddns-client/meta"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Root is the root command of the program
var Root = &cobra.Command{
	SilenceUsage: true,
	Use:          meta.ProgramFilename,
	Short:        "Update a CloudFlare DNS record with your public IP",
	Long: `A dynamic DNS client for CloudFlare. Automatically detects your public IP and
creates/updates a DNS record in CloudFlare.

Configuration flags can be set by defining an environment variable of the same name.
For example:
    DOMAIN=mydomain.com RECORD=sub.mydomain.com TOKEN=<api-token> cloudflare-ddns
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if conf.Daemon.Get() {
			return errors.Trace(ddns.DaemonWithDefaults(context.Background()))
		}
		return errors.Trace(ddns.Run(context.Background()))
	},
	Version: meta.Version,
}

func init() {
	if path, err := os.Executable(); err == nil {
		meta.ProgramDir = filepath.Dir(path)
		meta.ProgramFilename = filepath.Base(path)
	} else {
		panic(err)
	}
	f := Root.PersistentFlags()
	conf.Config.BindVar(f, &meta.ConfigFile).Viper()
	conf.Domain.Bind(f).Viper()
	conf.Record.Bind(f).Viper()
	conf.Token.Bind(f).Viper()
	conf.JSONOutput.Bind(f).Viper()
	conf.Verbose.Bind(f).Viper()
	conf.Daemon.Bind(f)
	Root.SetVersionTemplate("{{.Version}}\n")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if meta.ConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(meta.ConfigFile)
	} else {
		viper.AddConfigPath(meta.ProgramDir)
		viper.AddConfigPath("$HOME/.config")
		viper.AddConfigPath("/etc")
		viper.SetConfigName(meta.DefaultConfigFilename)
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Config file is optional, ignore errors
	err := viper.ReadInConfig()
	// TODO: use enums/string consts instead of hardcoded string "json"
	if conf.JSONOutput.Get() != "json" {
		writer := zerolog.ConsoleWriter{Out: os.Stderr}
		log.Logger = log.Output(writer)
	}
	if conf.Verbose.Get() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	// Config file is optional, continue if not found, unless config was specified by user and still not found
	_, notFound := err.(viper.ConfigFileNotFoundError)
	if !(notFound && meta.ConfigFile == "") {
		errhandler.Handle(err)
	}
}

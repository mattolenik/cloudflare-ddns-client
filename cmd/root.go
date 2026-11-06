package cmd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/conf"
	"github.com/mattolenik/cloudflare-ddns-client/ddns"
	"github.com/mattolenik/cloudflare-ddns-client/errhandler"
	"github.com/mattolenik/cloudflare-ddns-client/meta"
	"github.com/mattolenik/cloudflare-ddns-client/providers"
	"github.com/mattolenik/cloudflare-ddns-client/task"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Root is the root command of the program
var Root = &cobra.Command{
	SilenceUsage: true,
	Use:          "cloudflare-ddns",
	Short:        "Update a CloudFlare DNS record with your public IP",
	Long: `A dynamic DNS client for CloudFlare. Automatically detects your public IP and
creates/updates a DNS record in CloudFlare.

Configuration flags can be set by defining an environment variable of the same name.
For example:
    DOMAIN=mydomain.com RECORD=sub.mydomain.com TOKEN=<api-token> cloudflare-ddns
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := providers.NewCloudFlareProvider(context.Background(), conf.Token.Get())
		if err != nil {
			return errors.Annotatef(err, "failed to configure DDNS provider")
		}
		daemon := ddns.NewDefaultDaemon(provider, ddns.NewDefaultIPProvider(), ddns.NewDefaultConfigProvider())
		if conf.Daemon.Get() {
			return errors.Trace(runDaemon(provider, daemon))
		}
		return errors.Trace(daemon.Update())
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
	conf.Config.BindVar(f, &conf.ConfigFile)
	conf.Domain.Bind(f).WithDefault()
	conf.IP.Bind(f)
	conf.Record.Bind(f).WithDefault()
	conf.Token.Bind(f).WithDefault()
	conf.JSONOutput.Bind(f).WithDefault()
	conf.Verbose.Bind(f).WithDefault()
	conf.Daemon.Bind(f).WithDefault()
	Root.SetVersionTemplate("{{.Version}}\n")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	asus := conf.ConfigFile == conf.AsusNVRAMConfig
	var w io.Writer = os.Stderr
	// TODO: use enums/string consts instead of hardcoded string "json"
	if conf.JSONOutput.Get() != "json" {
		w = zerolog.ConsoleWriter{Out: os.Stderr}
	}
	if asus {
		// Also send logs to the router's system log
		w = zerolog.MultiLevelWriter(w, zerolog.ConsoleWriter{
			Out:          conf.AsusLogWriter{},
			NoColor:      true,
			PartsExclude: []string{zerolog.TimestampFieldName},
		})
	}
	log.Logger = log.Output(w)
	if conf.Verbose.Get() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if asus {
		log.Info().Msg("Using configuration from Asus router nvram")
		errhandler.Handle(conf.LoadAsusNVRAM())
		return
	}

	if conf.ConfigFile != "" {
		log.Info().Msgf("Using configuration from file '%s'", conf.ConfigFile)
		// Use config file from the flag.
		viper.SetConfigFile(conf.ConfigFile)
	} else {
		viper.AddConfigPath(meta.ProgramDir)
		viper.AddConfigPath("$HOME/.config")
		viper.AddConfigPath("/etc/")
		viper.SetConfigName(conf.DefaultConfigFilename)
	}

	// Config file is optional, continue if not found, unless config was specified by user and still not found
	err := viper.ReadInConfig()
	_, notFound := err.(viper.ConfigFileNotFoundError)
	if !(notFound && conf.ConfigFile == "") {
		errhandler.Handle(err)
	}
}

func runDaemon(provider ddns.DDNSProvider, daemon *ddns.DDNSDaemon) error {
	statusChan := daemon.StartWithDefaults()
	for {
		select {
		case status := <-statusChan:
			switch status.Type {
			case task.Info:
				log.Info().Msg(status.Message)
			case task.Error:
				log.Error().Msg(status.Message)
			case task.Fatal:
				log.Error().Msg("FATAL: " + status.Message)
				return status.Error
			}
			if status.IsDone {
				return nil
			}
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

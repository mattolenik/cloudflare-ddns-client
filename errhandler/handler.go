package errhandler

import (
	"os"
	"strings"

	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/conf"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Handle checks an error and exits if non-nil, printing a stack trace if applicable.
func Handle(err error) {
	if err == nil {
		return
	}
	msg := errors.ErrorStack(err)
	if conf.ModuleName != "" {
		// Remove name of module, makes stack traces shorter an easier to read
		msg = strings.ReplaceAll(msg, conf.ModuleName+"/", "")
	}
	if viper.GetBool(conf.JSONOutput) {
		// If writing to JSON logs, collapse stack trace into one line
		msg = strings.ReplaceAll(msg, "\n", " ↩ ")
	}
	log.Error().Msg(msg)
	os.Exit(1)
}

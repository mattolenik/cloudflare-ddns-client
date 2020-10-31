package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/lixiangzhong/dnsutil"
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

	ip, err := GetExternalIP()
	if err != nil {
		return errors.Annotate(err, "unable to retrieve external IP")
	}
	fmt.Println(ip)
	return nil
}

// GetExternalIP tries to detect the external IP address of this machine. First using DNS, then using several public HTTP APIs.
func GetExternalIP() (string, error) {
	ip, success, errs := GetExternalIPFromDNS()
	if !success && len(errs) > 0 {
		log.Warn().Msgf("failed to retrieve IP from DNS: %+v", errs)
		ip, success, errs := GetExternalIPFromAPIs()
		if !success && len(errs) > 0 {
			return "", errors.Errorf("failed to retrieve IP from public API: %+v", errs)
		} else if success && len(errs) > 0 {
			log.Warn().Msgf("successfully retrieved IP from public API but ran into the following problems: %+v", errs)
			return ip, nil
		} else if success {
			return ip, nil
		}
		// This shouldn't happen, if success is false, errs should always be len > 0
		return "", errors.New("unexpected error when retrieving IP with API, indicates a bug")
	} else if success && len(errs) > 0 {
		log.Warn().Msgf("successfully retrieved IP from DNS but ran into the following problems: %+v", errs)
		return ip, nil
	} else if success {
		return ip, nil
	}
	// This shouldn't happen, if success is false, errs should always be len > 0
	return "", errors.New("unexpected error when retrieving IP with DNS, indicates a bug")
}

// GetExternalIPFromDNS tries to detect the external IP address of this machine using DNS. First OpenDNS, then Google.
func GetExternalIPFromDNS() (string, bool, []error) {
	errs := []error{}
	success := false
	ip, err := GetIPFromOpenDNS()
	if err != nil {
		errs = append(errs, errors.Annotatef(err, "failed to retrieve IP from OpenDNS"))
		ip, err = GetIPFromGoogle()
		if err != nil {
			errs = append(errs, errors.Annotatef(err, "failed to retrieve IP from Google"))
		} else {
			success = true
		}
	} else {
		success = true
	}
	return ip, success, errs
}

// GetExternalIPFromAPIs tries to detect the external IP address of this machine using several public HTTP APIs.
func GetExternalIPFromAPIs() (string, bool, []error) {
	http.DefaultClient = &http.Client{
		Timeout: 5 * time.Second,
	}
	failures := []error{}
	httpURLs := []string{"http://whatismyip.akamai.com", "https://ipecho.net/plain", "https://wtfismyip.com/text"}
	for _, url := range httpURLs {
		ip, err := GetIPFromHTTP(url)
		if err != nil {
			failures = append(failures, errors.Trace(err))
		} else {
			return ip, true, failures
		}
	}
	return "", false, failures
}

// IsValidIP returns whether or not an IP address is valid
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// GetIPFromHTTP performs and HTTP GET and returns the body as a string
func GetIPFromHTTP(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Annotatef(err, "HTTP GET '%s' failed", url)
	}
	if resp.StatusCode != 200 {
		return "", errors.Annotatef(err, "HTTP GET '%s' failed with status %d", url, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	ip := strings.TrimSpace(string(body))
	if !IsValidIP(ip) {
		return "", errors.Errorf("IP address from '%s' is malformed", url)
	}
	return ip, nil
}

// GetIPFromGoogle gets the public IP from Google
func GetIPFromGoogle() (string, error) {
	dig := dnsutil.Dig{}
	dig.RemoteAddr = "ns1.google.com:53"
	out, err := dig.TXT("o-o.myaddr.l.google.com")
	if err != nil {
		return "", errors.Annotate(err, "unable to dig for TXT record from Google DNS")
	}
	if out == nil || len(out) != 1 {
		return "", errors.Errorf("expected to find only 1 DNS record, found %d", len(out))
	}
	if out[0].Txt == nil || len(out[0].Txt) != 1 {
		return "", errors.Errorf("expected to find only 1 TXT record, found %d", len(out))
	}
	return out[0].Txt[0], nil
}

// GetIPFromOpenDNS gets the public IP from OpenDNS
func GetIPFromOpenDNS() (string, error) {
	dig := dnsutil.Dig{}
	dig.RemoteAddr = "resolver1.opendns.com:53"
	out, err := dig.A("myip.opendns.com")
	if err != nil {
		return "", errors.Annotate(err, "unable to dig A record from OpenDNS")
	}
	if out == nil || len(out) != 1 {
		return "", errors.Errorf("expected to find only 1 DNS record, found %d", len(out))
	}
	if len(out) != 1 {
		return "", errors.Errorf("expected to find only 1 DNS record, found %d", len(out))
	}
	return out[0].A.String(), nil
}

// PrintVersion prints the program version and exits.
func PrintVersion() {
	fmt.Println(version)
	os.Exit(0)
}

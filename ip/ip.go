package ip

import (
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/lixiangzhong/dnsutil"
	"github.com/rs/zerolog/log"
)

// GetExternalIP tries to detect the external IP address of this machine. First using DNS, then using several public HTTP APIs.
func GetExternalIP() (string, error) {
	ip, success, errs := getExternalIPFromDNS()
	if !success && len(errs) > 0 {
		log.Warn().Msgf("failed to retrieve IP from DNS: %+v", errs)
		ip, success, errs := getExternalIPFromAPIs()
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

// getExternalIPFromDNS tries to detect the external IP address of this machine using DNS. First OpenDNS, then Google.
func getExternalIPFromDNS() (string, bool, []error) {
	errs := []error{}
	success := false
	ip, err := getIPFromOpenDNS()
	if err != nil {
		errs = append(errs, errors.Annotatef(err, "failed to retrieve IP from OpenDNS"))
		ip, err = getIPFromGoogle()
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

// getExternalIPFromAPIs tries to detect the external IP address of this machine using several public HTTP APIs.
func getExternalIPFromAPIs() (string, bool, []error) {
	http.DefaultClient = &http.Client{
		Timeout: 5 * time.Second,
	}
	failures := []error{}
	httpURLs := []string{"http://whatismyip.akamai.com", "https://ipecho.net/plain", "https://wtfismyip.com/text"}
	for _, url := range httpURLs {
		ip, err := getIPFromHTTP(url)
		if err != nil {
			failures = append(failures, errors.Trace(err))
		} else {
			return ip, true, failures
		}
	}
	return "", false, failures
}

// getIPFromHTTP performs and HTTP GET and returns the body as a string
func getIPFromHTTP(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Annotatef(err, "HTTP GET '%s' failed", url)
	}
	if resp.StatusCode != 200 {
		return "", errors.Annotatef(err, "HTTP GET '%s' failed with status %d", url, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	ip := strings.TrimSpace(string(body))
	if !isValidIP(ip) {
		return "", errors.Errorf("IP address from '%s' is malformed", url)
	}
	return ip, nil
}

// getIPFromGoogle gets the public IP from Google
func getIPFromGoogle() (string, error) {
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

// getIPFromOpenDNS gets the public IP from OpenDNS
func getIPFromOpenDNS() (string, error) {
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

// isValidIP returns whether or not an IP address is valid
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

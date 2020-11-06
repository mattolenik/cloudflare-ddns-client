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

var apiURLs = []string{"http://whatismyip.akamai.com", "https://ipecho.net/plain", "https://wtfismyip.com/text"}

type dnsLookup struct {
	// DNS server to dig
	Address string
	// Name of record
	RecordName string
	// Type of record, either A or TXT
	RecordType string
}

var dnsLookupGoogle = dnsLookup{
	Address:    "ns1.google.com:53",
	RecordName: "o-o.myaddr.l.google.com",
	RecordType: "TXT",
}
var dnsLookupOpenDNS = dnsLookup{
	Address:    "resolver1.opendns.com:53",
	RecordName: "myip.opendns.com",
	RecordType: "A",
}

// GetPublicIPWithRetry calls GetPublicIP and with numRetries attempts waiting delayInSeconds after each attempt.
func GetPublicIPWithRetry(numRetries int, delay time.Duration) (string, error) {
	var i int
	for i = 0; i < numRetries; i++ {
		ip, err := GetPublicIP()
		if err == nil {
			return ip, nil
		}
		log.Warn().Msgf("failed to retrieve public IP, attempt #%d, retrying in %d", i+1, delay.String())
		time.Sleep(delay)
	}
	return "", errors.Errorf("failed to retrieve public IP after %d attempts", i)
}

// GetPublicIP tries to detect the public IP address of this machine. First using DNS, then using several public HTTP APIs.
func GetPublicIP() (string, error) {
	ip, success, errs := getPublicIPFromDNS(dnsLookupOpenDNS, dnsLookupGoogle)
	if !success && len(errs) > 0 {
		log.Warn().Msgf("failed to retrieve IP from DNS: %+v", errs)
		ip, success, errs := getPublicIPFromAPIs(apiURLs...)
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

// getPublicIPFromDNS tries to detect the public IP address of this machine using DNS. First OpenDNS, then Google.
func getPublicIPFromDNS(lookups ...dnsLookup) (string, bool, []error) {
	if len(lookups) == 0 {
		return "", false, []error{errors.New("expected at least one DNS lookup request")}
	}
	errs := []error{}
	for _, lookup := range lookups {
		var record string
		var err error
		switch lookup.RecordType {
		case "A":
			record, err = getDNSARecord(lookup.Address, lookup.RecordName)
		case "TXT":
			record, err = getDNSTXTRecord(lookup.Address, lookup.RecordName)
		default:
			return "", false, []error{errors.Errorf("unsupported record type '%s'", lookup.RecordType)}
		}
		if err != nil {
			errs = append(errs, errors.Annotatef(err, "DNS lookup of record '%s' at '%s' failed", lookup.RecordName, lookup.Address))
			continue
		}
		if !isValidIP(record) {
			errs = append(errs, errors.Errorf("invalid IP '%s' from DNS lookup of record '%s' at '%s'", record, lookup.RecordName, lookup.Address))
		}
		return record, true, nil
	}
	return "", false, errs
}

// getPublicIPFromAPIs tries to detect the public IP address of this machine using several public HTTP APIs.
// returns IP, success/fail, and one or more errors (one per API used)
func getPublicIPFromAPIs(apiURLs ...string) (string, bool, []error) {
	http.DefaultClient = &http.Client{
		Timeout: 5 * time.Second,
	}
	failures := []error{}
	for _, url := range apiURLs {
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
		return "", errors.Errorf("HTTP GET '%s' failed with status %s", url, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Annotatef(err, "failed reading response body from '%s'", url)
	}
	ip := strings.TrimSpace(string(body))
	if !isValidIP(ip) {
		return "", errors.Errorf("IP address from '%s' is malformed", url)
	}
	return ip, nil
}

// getDNSTXTRecord gets the public IP from a DNS TXT record
func getDNSTXTRecord(addr, record string) (string, error) {
	dig := dnsutil.Dig{}
	dig.RemoteAddr = addr
	out, err := dig.TXT(record)
	if err != nil {
		return "", errors.Annotatef(err, "unable to dig for TXT record '%s' from '%s'", record, addr)
	}
	if out == nil || len(out) != 1 {
		return "", errors.Errorf("expected to find only 1 DNS record, found %d", len(out))
	}
	if out[0].Txt == nil || len(out[0].Txt) != 1 {
		return "", errors.Errorf("expected to find only 1 TXT record, found %d", len(out))
	}
	return out[0].Txt[0], nil
}

//  getDNSARecord gets a DNS A record
func getDNSARecord(address, record string) (string, error) {
	dig := dnsutil.Dig{}
	dig.RemoteAddr = address
	out, err := dig.A(record)
	if err != nil {
		return "", errors.Annotatef(err, "unable to dig A record '%s' from '%s'", record, address)
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

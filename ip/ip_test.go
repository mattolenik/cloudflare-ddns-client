package ip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExternalIP(t *testing.T) {
	assert := assert.New(t)
	ip, err := GetExternalIP()
	assert.Truef(isValidIP(ip), "expected IP '%s' to be valid", ip)
	assert.NoErrorf(err, "expected no error when getting external IP")
}

func Test_getIPFromHTTP_MalformedResponse(t *testing.T) {
	assert := assert.New(t)
	ip, err := getIPFromHTTP("http://example.com/notanip")
	assert.Errorf(err, "expected error for malformed response")
	assert.Falsef(isValidIP(ip), "expected IP '%s' to be invalid", ip)
}

func TestIPFromHTTPAPIs(t *testing.T) {
	assert := assert.New(t)
	ips := []string{}
	for _, url := range apiURLs {
		ip, success, err := getExternalIPFromAPIs(url)
		assert.Emptyf(err, "expected no errors from API with URL '%s'", url)
		assert.Truef(success, "expected to get IP from API with URL '%s'", url)
		ips = append(ips, ip)
	}
	for _, ip := range ips {
		assert.Equalf(ip, ips[0], "expected all IPs from all IPs to be the same, but found '%+v'", ips)
	}
}

func TestIPFromHTTPAPIs_MalformedResponse(t *testing.T) {
	assert := assert.New(t)
	testURLs := []string{"http://example.com/notanip", apiURLs[1], apiURLs[2]}
	ip, success, errs := getExternalIPFromAPIs(testURLs...)
	assert.Truef(success, "expected to get IP despite one malformed response")
	assert.Truef(isValidIP(ip), "expected IP '%s' to be valid", ip)
	assert.Lenf(errs, 1, "expected exactly one error")
}

func TestIPFromHTTPAPIs_InvalidURL(t *testing.T) {
	assert := assert.New(t)
	testURLs := []string{"sfkuer", apiURLs[1], apiURLs[2]}
	ip, success, errs := getExternalIPFromAPIs(testURLs...)
	assert.Truef(success, "expected to get IP despite one invalid URL")
	assert.Truef(isValidIP(ip), "expected IP '%s' to be valid", ip)
	assert.Lenf(errs, 1, "expected exactly one error")
}

func TestIPFromHTTPAPIs_HostNotFound(t *testing.T) {
	assert := assert.New(t)
	testURLs := []string{"http://somethingthatdoesntexist55701230950.com", apiURLs[1], apiURLs[2]}
	ip, success, errs := getExternalIPFromAPIs(testURLs...)
	assert.Truef(success, "expected to get IP despite one host not found")
	assert.Truef(isValidIP(ip), "expected IP '%s' to be valid", ip)
	assert.Lenf(errs, 1, "expected exactly one error")
}

func TestIPFromDNS(t *testing.T) {
	assert := assert.New(t)

	ipOpenDNS, err := getIPFromOpenDNS()
	assert.NoErrorf(err, "expected no error getting IP from OpenDNS")
	assert.Truef(isValidIP(ipOpenDNS), "expected IP from OpenDNS '%s' to be valid", ipOpenDNS)

	ipGoogle, err := getIPFromGoogle()
	assert.NoErrorf(err, "expected no error getting IP from Google")
	assert.Truef(isValidIP(ipGoogle), "expected IP from Google '%s' to be valid", ipGoogle)
	assert.Equalf(ipOpenDNS, ipGoogle, "expected IPs from OpenDNS and Google to agree")
}

func TestIPFromDNSAndAPIsAgree(t *testing.T) {
	assert := assert.New(t)

	ipDNS, success, errs := getExternalIPFromDNS()
	assert.Truef(success, "expected success getting IP from DNS")
	assert.Lenf(errs, 0, "expected no errors from DNS")

	ipAPI, success, errs := getExternalIPFromAPIs(apiURLs...)
	assert.Truef(success, "expected success getting IP from HTTP APIs")
	assert.Lenf(errs, 0, "expected no errors from HTTP APIs")

	assert.Equalf(ipDNS, ipAPI, "expected IP from DNS '%s' to equal IP from APIs '%s'", ipDNS, ipAPI)
	assert.Truef(isValidIP(ipDNS), "expected IP '%s' to be valid", ipDNS)
}

func Test_isValidIP(t *testing.T) {
	assert := assert.New(t)
	assert.False(isValidIP(""))
	assert.True(isValidIP("192.168.0.1"))
	assert.False(isValidIP("<html><body>some invalid response</body></html>"))
}

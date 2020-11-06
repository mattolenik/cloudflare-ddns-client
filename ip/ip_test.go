package ip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPublicIP(t *testing.T) {
	assert := assert.New(t)
	ip, err := GetPublicIP()
	assert.Truef(isValidIP(ip), "expected IP '%s' to be valid", ip)
	assert.NoErrorf(err, "expected no error when getting public IP")
}

func TestIPFromHTTPAPIs(t *testing.T) {
	assert := assert.New(t)
	ips := []string{}
	for _, url := range apiURLs {
		ip, success, err := getPublicIPFromAPIs(url)
		assert.Emptyf(err, "expected no errors from API with URL '%s'", url)
		assert.Truef(success, "expected to get IP from API with URL '%s'", url)
		ips = append(ips, ip)
	}
	for _, ip := range ips {
		assert.Equalf(ip, ips[0], "expected all IPs from all APIs to be the same, but found '%+v'", ips)
	}
}

func TestIPFromHTTPAPIs_MalformedResponse(t *testing.T) {
	assert := assert.New(t)
	testURLs := []string{"http://example.com/notanip", apiURLs[1], apiURLs[2]}
	ip, success, errs := getPublicIPFromAPIs(testURLs...)
	assert.Truef(success, "expected to get IP despite one malformed response")
	assert.Truef(isValidIP(ip), "expected IP '%s' to be valid", ip)
	assert.Lenf(errs, 1, "expected exactly one error")
}

func TestIPFromHTTPAPIs_InvalidURL(t *testing.T) {
	assert := assert.New(t)
	testURLs := []string{"sfkuer", apiURLs[1], apiURLs[2]}
	ip, success, errs := getPublicIPFromAPIs(testURLs...)
	assert.Truef(success, "expected to get IP despite one invalid URL")
	assert.Truef(isValidIP(ip), "expected IP '%s' to be valid", ip)
	assert.Lenf(errs, 1, "expected exactly one error")
}

func TestIPFromHTTPAPIs_HostNotFound(t *testing.T) {
	assert := assert.New(t)
	testURLs := []string{"http://somethingthatdoesntexist55701230950.com", apiURLs[1], apiURLs[2]}
	ip, success, errs := getPublicIPFromAPIs(testURLs...)
	assert.Truef(success, "expected to get IP despite one host not found")
	assert.Truef(isValidIP(ip), "expected IP '%s' to be valid", ip)
	assert.Lenf(errs, 1, "expected exactly one error")
}

func TestIPFromDNS(t *testing.T) {
	assert := assert.New(t)
	ipOpenDNS, success, errs := getPublicIPFromDNS(dnsLookupOpenDNS)
	assert.Truef(success, "expected success getting IP from OpenDNS")
	assert.Emptyf(errs, "expected no error getting IP from OpenDNS")
	assert.Truef(isValidIP(ipOpenDNS), "expected IP from OpenDNS '%s' to be valid", ipOpenDNS)

	ipGoogle, success, errs := getPublicIPFromDNS(dnsLookupGoogle)
	assert.Truef(success, "expected success getting IP from Google")
	assert.Emptyf(errs, "expected no error getting IP from Google")
	assert.Truef(isValidIP(ipOpenDNS), "expected IP from  Google'%s' to be valid", ipGoogle)
}

func TestIPFromDNS_WithFailure(t *testing.T) {
	assert := assert.New(t)
	// Ensure that DNS lookup works despite multiple failures
	ipOpenDNS, success, errs := getPublicIPFromDNS(
		// Invalid DNS server case
		dnsLookup{
			Address:    "ns1.invaliddnsdoesnotexist:53",
			RecordName: "invalidrecord",
			RecordType: "A",
		},
		// Valid DNS server but invalid record
		dnsLookup{
			Address:    dnsLookupOpenDNS.Address,
			RecordName: "invalidrecord",
			RecordType: "A",
		},
		// Valid DNS server and record but incorrect type
		dnsLookup{
			Address:    dnsLookupOpenDNS.Address,
			RecordName: dnsLookupOpenDNS.RecordName,
			RecordType: "TXT",
		},
		dnsLookupOpenDNS,
	)
	assert.Truef(success, "expected success getting IP from OpenDNS")
	assert.Emptyf(errs, "expected no error getting IP from OpenDNS")
	assert.Truef(isValidIP(ipOpenDNS), "expected IP from OpenDNS '%s' to be valid", ipOpenDNS)

	ipGoogle, success, errs := getPublicIPFromDNS(dnsLookupGoogle)
	assert.Truef(success, "expected success getting IP from Google")
	assert.Emptyf(errs, "expected no error getting IP from Google")
	assert.Truef(isValidIP(ipOpenDNS), "expected IP from  Google'%s' to be valid", ipGoogle)
}

func TestIPFromDNSAndAPIsAgree(t *testing.T) {
	assert := assert.New(t)

	ipDNS, success, errs := getPublicIPFromDNS(dnsLookupOpenDNS, dnsLookupGoogle)
	assert.Truef(success, "expected success getting IP from DNS")
	assert.Lenf(errs, 0, "expected no errors from DNS")

	ipAPI, success, errs := getPublicIPFromAPIs(apiURLs...)
	assert.Truef(success, "expected success getting IP from HTTP APIs")
	assert.Lenf(errs, 0, "expected no errors from HTTP APIs")

	assert.Equalf(ipDNS, ipAPI, "expected IP from DNS '%s' to equal IP from APIs '%s'", ipDNS, ipAPI)
	assert.Truef(isValidIP(ipDNS), "expected IP '%s' to be valid", ipDNS)
}

func Test_getIPFromHTTP_MalformedResponse(t *testing.T) {
	assert := assert.New(t)
	ip, err := getIPFromHTTP("http://example.com/notanip")
	assert.Errorf(err, "expected error for malformed response")
	assert.Falsef(isValidIP(ip), "expected IP '%s' to be invalid", ip)
}

func TestDNSLookupFailures(t *testing.T) {
	assert := assert.New(t)

	_, err := getDNSTXTRecord("ns1.invaliddnsdoesnotexist.com:53", "o-o.myaddr.l.google.com")
	assert.Errorf(err, "expected DNS lookup to fail due to invalid address")

	_, err = getDNSTXTRecord("ns1.google.com:53", "invalidrecord")
	assert.Errorf(err, "expected DNS lookup to fail due to invalid record")

	_, err = getDNSARecord("resolver1.invaliddnsdoesnotexist.com:53", "myip.opendns.com")
	assert.Errorf(err, "expected DNS lookup to fail due to invalid address")

	_, err = getDNSARecord("resolver1.opendns.com:53", "invalidrecord")
	assert.Errorf(err, "expected DNS lookup to fail due to invalid record")
}

func Test_isValidIP(t *testing.T) {
	assert := assert.New(t)
	assert.False(isValidIP(""))
	assert.True(isValidIP("192.168.0.1"))
	assert.False(isValidIP("<html><body>some invalid response</body></html>"))
}

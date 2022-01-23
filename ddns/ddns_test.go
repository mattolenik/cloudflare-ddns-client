package ddns

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mattolenik/cloudflare-ddns-client/ddns"
	"github.com/mattolenik/cloudflare-ddns-client/test"
)

func TestUpdate(t *testing.T) {
	assert, _, ctrl, cleanup := test.NewTools(t)
	defer cleanup()

	domain := "abc.com"
	record := "xyz.abc.com"
	ip := "1.1.1.1"

	ddnsProvider := NewMockDDNSProvider(ctrl)
	ipProvider := NewMockIPProvider(ctrl)
	configProvider := NewMockConfigProvider(ctrl)
	ddnsDaemon := NewDefaultDaemon(ddnsProvider, ipProvider, configProvider)

	configProvider.EXPECT().Get().Return(domain, record, nil)
	ipProvider.EXPECT().Get().Return(ip, nil)
	ddnsProvider.EXPECT().Update(gomock.Eq(domain), gomock.Eq(record), gomock.Eq(ip)).Return(nil).Times(1)
	ddnsProvider.EXPECT().Get(domain, record).Return(ip, nil).Times(1)
	assert.NoError(ddnsDaemon.Update())

	actualIP, err := ddnsProvider.Get(domain, record)
	assert.NoError(err)
	assert.Equal(ip, actualIP)
}

func TestDaemon(t *testing.T) {
	updatePeriod := 50 * time.Millisecond
	retryDelay := 50 * time.Millisecond
	domain := "abc.com"
	record := "xyz.abc.com"
	currentSuffix := 0
	currentIP := ""
	ipGen := func() (string, error) {
		currentSuffix++
		currentIP = fmt.Sprintf("1.1.1.%d", currentSuffix)
		return currentIP, nil
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ddnsProvider, ipProvider, configProvider := fixtures(ctrl)
	ddnsDaemon := ddns.NewDefaultDaemon(ddnsProvider, ipProvider, configProvider)

	getCurrentIP := func() interface{} { return currentIP }

	ddnsDaemon.Start(updatePeriod, retryDelay)

	configProvider.EXPECT().Get().Return(domain, record, nil)
	ipProvider.EXPECT().Get().DoAndReturn(ipGen)
	//ddnsProvider.EXPECT().Update(gomock.Eq(domain), gomock.Eq(record), gomock.Eq(ip)).Return(nil).Times(1)
	//ddnsProvider.EXPECT().Get(domain, record).Return(ip, nil).Times(1)
	ddnsProvider.EXPECT().
		Update(
			gomock.Eq(domain),
			gomock.Eq(record),
			FnMatch(gomock.Eq, getCurrentIP),
		).AnyTimes()
}

func fixtures(ctrl *gomock.Controller) (ddnsProvider *MockDDNSProvider, ipProvider *MockIPProvider, configProvider *MockConfigProvider) {
	return NewMockDDNSProvider(ctrl),
		NewMockIPProvider(ctrl),
		NewMockConfigProvider(ctrl)
}

type funcMatcher struct {
	gomock.Matcher
	value   func() interface{}
	matchFn func(interface{}) gomock.Matcher
}

func FnMatch(matchFn func(interface{}) gomock.Matcher, value func() interface{}) gomock.Matcher {
	return &funcMatcher{
		matchFn: matchFn,
		value:   value,
	}
}

func (f *funcMatcher) Matches(x interface{}) bool {
	return f.matchFn(x).Matches(x)
}

func (f *funcMatcher) String() string {
	return "runs underlying match against new value each time"
}

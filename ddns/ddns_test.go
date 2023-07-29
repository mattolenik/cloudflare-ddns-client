package ddns

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mattolenik/cloudflare-ddns-client/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type TestFlow struct {
	t           *testing.T
	assert      *assert.Assertions
	require     *require.Assertions
	description string
}

func It(t *testing.T, desc string) *TestFlow {
	return &TestFlow{
		description: desc,
		t:           t,
		assert:      assert.New(t),
		require:     require.New(t),
	}
}

// TODO: Make generic so should funcs can take any function type? Make not dependent upon testify if possible
type ShouldFunc func(t *testing.T, assert *assert.Assertions, require *require.Assertions)

func (f *TestFlow) Should(msg string, fn ShouldFunc) *TestFlow {
	f.t.Run(fmt.Sprintf("It '%s' should '%s'", f.description, msg), func(t *testing.T) { fn(f.t, f.assert, f.require) })
	return f
}

func TestDaemon(t *testing.T) {
	daemon := It(t, "daemon")

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
	ddnsDaemon := NewDefaultDaemon(ddnsProvider, ipProvider, configProvider)

	daemon.Should("run", func(t *testing.T, assert *assert.Assertions, require *require.Assertions) {
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
	})
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

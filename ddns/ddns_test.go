package ddns

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mattolenik/cloudflare-ddns-client/task"
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
	assert, require, ctrl, cleanup := test.NewTools(t)
	defer cleanup()
	ddnsProvider, ipProvider, configProvider := fixtures(ctrl)
	d := NewDefaultDaemon(ddnsProvider, ipProvider, configProvider)

	domain, record := "abc.com", "xyz.abc.com"
	// Public IP stays the same for two checks, then changes and stays changed
	publicIPs := []string{"1.1.1.1", "1.1.1.1", "2.2.2.2"}
	recordIP := ""
	configProvider.EXPECT().Get().Return(domain, record, nil).AnyTimes()
	ipProvider.EXPECT().Get().DoAndReturn(func() (string, error) {
		ip := publicIPs[0]
		if len(publicIPs) > 1 {
			publicIPs = publicIPs[1:]
		}
		return ip, nil
	}).AnyTimes()
	ddnsProvider.EXPECT().Get(domain, record).DoAndReturn(func(_, _ string) (string, error) { return recordIP, nil }).AnyTimes()
	setRecord := func(_, _, ip string) error { recordIP = ip; return nil }
	gomock.InOrder(
		ddnsProvider.EXPECT().Update(domain, record, "1.1.1.1").DoAndReturn(setRecord),
		ddnsProvider.EXPECT().Update(domain, record, "2.2.2.2").DoAndReturn(setRecord),
	)

	statuses := d.Start(10*time.Millisecond, 10*time.Millisecond)
	next := func() string {
		select {
		case s, ok := <-statuses:
			require.True(ok, "status channel closed early")
			assert.Equal(task.Info, s.Type, s.Message)
			return s.Message
		case <-time.After(time.Second):
			require.FailNow("timed out waiting for daemon status")
			return ""
		}
	}
	assert.Contains(next(), "Daemon running")
	assert.Equal("Found public IP '1.1.1.1'", next())
	assert.Contains(next(), "No IP change detected")
	assert.Equal("Detected new public IP address, it changed from '1.1.1.1' to '2.2.2.2'", next())
	assert.Contains(next(), "No IP change detected")

	d.Stop()
	done := make(chan struct{})
	go func() {
		for range statuses {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		require.FailNow("daemon did not stop")
	}
}

func fixtures(ctrl *gomock.Controller) (ddnsProvider *MockDDNSProvider, ipProvider *MockIPProvider, configProvider *MockConfigProvider) {
	return NewMockDDNSProvider(ctrl),
		NewMockIPProvider(ctrl),
		NewMockConfigProvider(ctrl)
}

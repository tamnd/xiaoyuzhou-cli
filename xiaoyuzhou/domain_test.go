package xiaoyuzhou

import (
	"testing"
)

// These tests are offline: they exercise the domain metadata and kit wiring.
// Network behaviour is covered in xiaoyuzhou_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "xiaoyuzhou" {
		t.Errorf("Scheme = %q, want xiaoyuzhou", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "xiaoyuzhou" {
		t.Errorf("Identity.Binary = %q, want xiaoyuzhou", info.Identity.Binary)
	}
}

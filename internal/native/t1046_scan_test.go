//go:build windows

package native

import (
	"testing"
)

func TestSubnetHosts_ValidCIDR(t *testing.T) {
	hosts, err := subnetHosts("192.168.1.0/24")
	if err != nil {
		t.Fatalf("subnetHosts() returned unexpected error: %v", err)
	}
	if len(hosts) != 254 {
		t.Errorf("subnetHosts() returned %d hosts, want 254", len(hosts))
	}
	if hosts[0] != "192.168.1.1" {
		t.Errorf("subnetHosts() first host = %q, want %q", hosts[0], "192.168.1.1")
	}
	if hosts[len(hosts)-1] != "192.168.1.254" {
		t.Errorf("subnetHosts() last host = %q, want %q", hosts[len(hosts)-1], "192.168.1.254")
	}
}

func TestSubnetHosts_ExcludesOwnIP(t *testing.T) {
	hosts, err := subnetHosts("192.168.1.0/24", "192.168.1.50")
	if err != nil {
		t.Fatalf("subnetHosts() returned unexpected error: %v", err)
	}
	if len(hosts) != 253 {
		t.Errorf("subnetHosts() returned %d hosts, want 253", len(hosts))
	}
	for _, h := range hosts {
		if h == "192.168.1.50" {
			t.Error("subnetHosts() result contains excluded IP 192.168.1.50")
		}
	}
}

func TestSubnetHosts_ExcludesNetworkAndBroadcast(t *testing.T) {
	hosts, err := subnetHosts("192.168.1.0/24")
	if err != nil {
		t.Fatalf("subnetHosts() returned unexpected error: %v", err)
	}
	for _, h := range hosts {
		if h == "192.168.1.0" {
			t.Error("subnetHosts() result contains network address 192.168.1.0")
		}
		if h == "192.168.1.255" {
			t.Error("subnetHosts() result contains broadcast address 192.168.1.255")
		}
	}
}

func TestSubnetHosts_InvalidCIDR(t *testing.T) {
	_, err := subnetHosts("unknown")
	if err == nil {
		t.Error("subnetHosts(\"unknown\") expected error, got nil")
	}
}

func TestLocalIPAddrs(t *testing.T) {
	addrs := localIPAddrs()
	if len(addrs) < 1 {
		t.Error("localIPAddrs() returned empty slice, want at least 1 IP")
	}
}

func TestTCPPortList(t *testing.T) {
	if len(tcpPorts) != 17 {
		t.Errorf("tcpPorts has %d entries, want 17", len(tcpPorts))
	}
	expected := []int{21, 22, 23, 25, 53, 80, 135, 139, 389, 443, 445, 1433, 3306, 3389, 5985, 8080, 8443}
	for i, p := range expected {
		if tcpPorts[i] != p {
			t.Errorf("tcpPorts[%d] = %d, want %d", i, tcpPorts[i], p)
		}
	}
}

func TestUDPPortList(t *testing.T) {
	if len(udpPorts) != 3 {
		t.Errorf("udpPorts has %d entries, want 3", len(udpPorts))
	}
	expected := []int{53, 161, 123}
	for i, p := range expected {
		if udpPorts[i] != p {
			t.Errorf("udpPorts[%d] = %d, want %d", i, udpPorts[i], p)
		}
	}
}

func TestT1046Registered(t *testing.T) {
	fn := Lookup("T1046")
	if fn == nil {
		t.Error("Lookup(\"T1046\") returned nil — T1046 not registered in native registry")
	}
}

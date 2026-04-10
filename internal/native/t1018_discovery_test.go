//go:build windows

package native

import (
	"testing"
)

func TestParseARPTable_ValidOutput(t *testing.T) {
	input := "Interface: 192.168.1.100 --- 0x5\n  Internet Address      Physical Address      Type\n  192.168.1.1           aa-bb-cc-dd-ee-ff     dynamic\n  192.168.1.50          11-22-33-44-55-66     dynamic\n"
	entries := parseARPTable(input)
	if len(entries) != 2 {
		t.Fatalf("parseARPTable() returned %d entries, want 2", len(entries))
	}
	if entries[0].IP != "192.168.1.1" {
		t.Errorf("entries[0].IP = %q, want %q", entries[0].IP, "192.168.1.1")
	}
	if entries[0].MAC != "aa-bb-cc-dd-ee-ff" {
		t.Errorf("entries[0].MAC = %q, want %q", entries[0].MAC, "aa-bb-cc-dd-ee-ff")
	}
	if entries[1].IP != "192.168.1.50" {
		t.Errorf("entries[1].IP = %q, want %q", entries[1].IP, "192.168.1.50")
	}
	if entries[1].MAC != "11-22-33-44-55-66" {
		t.Errorf("entries[1].MAC = %q, want %q", entries[1].MAC, "11-22-33-44-55-66")
	}
}

func TestParseARPTable_SkipsHeaders(t *testing.T) {
	input := "Interface: 192.168.1.100 --- 0x5\n  Internet Address      Physical Address      Type\n  192.168.1.1           aa-bb-cc-dd-ee-ff     dynamic\n"
	entries := parseARPTable(input)
	if len(entries) != 1 {
		t.Fatalf("parseARPTable() returned %d entries, want 1 (headers should be skipped)", len(entries))
	}
}

func TestParseARPTable_EmptyOutput(t *testing.T) {
	entries := parseARPTable("")
	if len(entries) != 0 {
		t.Errorf("parseARPTable(\"\") returned %d entries, want 0", len(entries))
	}
}

func TestNltestDCDiscovery_Runs(t *testing.T) {
	result := nltestDCDiscovery()
	if result == "" {
		t.Error("nltestDCDiscovery() returned empty string — expected a non-empty string (graceful fallback)")
	}
}

func TestDnsReverseLookup_Localhost(t *testing.T) {
	result := dnsReverseLookup([]string{"127.0.0.1"})
	// Result may be empty if no PTR record configured, but must not panic
	_ = result
}

func TestIcmpChecksum(t *testing.T) {
	// ICMP echo request header: type=8, code=0, checksum=0 (placeholder), id=0, seq=1
	msg := []byte{8, 0, 0, 0, 0, 0, 0, 1}
	cs := icmpChecksum(msg)
	// checksum of this 8-byte message: sum of 16-bit words
	// 0x0800 + 0x0000 + 0x0000 + 0x0001 = 0x0801 -> ~0x0801 = 0xF7FE
	want := uint16(0xF7FE)
	if cs != want {
		t.Errorf("icmpChecksum() = 0x%04X, want 0x%04X", cs, want)
	}
}

func TestT1018Registered(t *testing.T) {
	fn := Lookup("T1018")
	if fn == nil {
		t.Error("Lookup(\"T1018\") returned nil — T1018 not registered in native registry")
	}
}

//go:build windows

package native

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// arpEntry holds a parsed entry from `arp -a` output.
type arpEntry struct {
	IP  string
	MAC string
}

// icmpChecksum computes the RFC 1071 Internet checksum of data.
// Sums all 16-bit words, folds carry bits, and returns one's complement.
func icmpChecksum(data []byte) uint16 {
	var sum uint32
	length := len(data)
	for i := 0; i+1 < length; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	// Handle odd byte
	if length%2 != 0 {
		sum += uint32(data[length-1]) << 8
	}
	// Fold 32-bit sum into 16 bits
	for sum>>16 != 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	return ^uint16(sum)
}

// icmpPingSweep sends ICMP echo requests to each host and returns the slice of
// hosts that respond. Requires raw socket privilege (admin). If ListenPacket
// fails (non-admin), falls back to tcpAliveCheck.
func icmpPingSweep(hosts []string) []string {
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		// Non-admin path: fall back to TCP alive check per D-09
		return tcpAliveCheck(hosts)
	}
	defer conn.Close()

	var alive []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, scanWorkers)

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Build ICMP echo request: type=8, code=0, checksum(placeholder), id=0, seq=1
			msg := []byte{8, 0, 0, 0, 0, 0, 0, 1}
			cs := icmpChecksum(msg)
			msg[2] = byte(cs >> 8)
			msg[3] = byte(cs)

			dst, err := net.ResolveIPAddr("ip4", h)
			if err != nil {
				return
			}

			if err := conn.SetDeadline(time.Now().Add(300 * time.Millisecond)); err != nil {
				return
			}
			if _, err := conn.WriteTo(msg, dst); err != nil {
				return
			}

			buf := make([]byte, 64)
			conn.SetDeadline(time.Now().Add(300 * time.Millisecond)) //nolint:errcheck
			n, addr, err := conn.ReadFrom(buf)
			if err != nil || n < 1 {
				return
			}
			// Check for ICMP echo reply (type=0) — IP header is 20 bytes, ICMP starts at offset 20
			if addr.String() == h && n >= 21 && buf[20] == 0 {
				mu.Lock()
				alive = append(alive, h)
				mu.Unlock()
			}
		}(host)
	}

	wg.Wait()
	return alive
}

// tcpAliveCheck tests whether each host is alive by attempting TCP connections
// to ports 445 and 135. Uses a bounded goroutine pool (scanWorkers) for speed.
// This is the fallback path when ICMP requires admin privileges (D-09).
func tcpAliveCheck(hosts []string) []string {
	var alive []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, scanWorkers)

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			for _, port := range []string{"445", "135"} {
				conn, err := net.DialTimeout("tcp4", h+":"+port, 300*time.Millisecond)
				if err == nil {
					conn.Close()
					mu.Lock()
					alive = append(alive, h)
					mu.Unlock()
					return
				}
			}
		}(host)
	}

	wg.Wait()
	return alive
}

// parseARPTable parses the output of `arp -a` and returns a slice of arpEntry.
// Lines where the first whitespace-separated field is not a valid IP are skipped
// (this excludes header lines like "Interface: ..." and column headers).
func parseARPTable(output string) []arpEntry {
	var entries []arpEntry
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ip := net.ParseIP(fields[0])
		if ip == nil {
			continue
		}
		entries = append(entries, arpEntry{IP: fields[0], MAC: fields[1]})
	}
	return entries
}

// arpTableDump executes `arp -a` and returns the raw output string plus parsed entries.
func arpTableDump() (string, []arpEntry) {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return fmt.Sprintf("arp -a failed: %s\n", err.Error()), nil
	}
	raw := string(out)
	return raw, parseARPTable(raw)
}

// nltestDCDiscovery runs nltest.exe to discover domain controllers.
// Falls back gracefully when the machine is not domain-joined or nltest is unavailable.
func nltestDCDiscovery() string {
	domain := os.Getenv("USERDNSDOMAIN")
	if domain == "" {
		domain = "."
	}

	// Try nltest.exe in PATH first, then fall back to full system path
	for _, cmd := range []string{"nltest.exe", `C:\Windows\System32\nltest.exe`} {
		out, err := exec.Command(cmd, "/dsgetdc:"+domain).CombinedOutput()
		if err != nil {
			// Check if the binary was found at all (exec error vs. non-zero exit)
			if strings.Contains(err.Error(), "executable file not found") ||
				strings.Contains(err.Error(), "The system cannot find") {
				continue
			}
			// Binary found but returned non-zero: not domain-joined
			return "nltest: DC not found (non-domain environment)\n"
		}
		return string(out)
	}

	return "nltest: not available (non-domain environment or tool not found)\n"
}

// dnsReverseLookup performs PTR lookups on each host IP and returns formatted results.
// Hosts with no PTR record are silently skipped.
func dnsReverseLookup(hosts []string) string {
	var sb strings.Builder
	for _, ip := range hosts {
		names, err := net.LookupAddr(ip)
		if err != nil || len(names) == 0 {
			continue
		}
		fmt.Fprintf(&sb, "  PTR: %s -> %s\n", ip, strings.Join(names, ", "))
	}
	return sb.String()
}

// runT1018 executes T1018 Remote System Discovery by chaining four methods:
//  1. ICMP ping sweep (admin) or TCP 445/135 alive check (non-admin)
//  2. ARP table dump via `arp -a`
//  3. Domain controller discovery via nltest.exe
//  4. DNS reverse lookups on alive hosts (union of ICMP + ARP IPs)
func runT1018() (NativeResult, error) {
	subnet := localSubnet()
	if subnet == "unknown" {
		return NativeResult{
			Success:     false,
			ErrorOutput: "T1018: cannot detect local subnet — no active non-loopback IPv4 interface found",
		}, fmt.Errorf("T1018: cannot detect local subnet")
	}

	ownIPs := localIPAddrs()
	hosts, err := subnetHosts(subnet, ownIPs...)
	if err != nil {
		return NativeResult{
			Success:     false,
			ErrorOutput: fmt.Sprintf("T1018: subnet enumeration failed: %s", err.Error()),
		}, fmt.Errorf("T1018 subnetHosts: %w", err)
	}

	var sb strings.Builder

	// === Step 1: ICMP Ping Sweep (or TCP fallback) ===
	fmt.Fprintf(&sb, "=== ICMP Ping Sweep ===\n")
	fmt.Fprintf(&sb, "Scanning %s (%d hosts)...\n", subnet, len(hosts))
	aliveFromICMP := icmpPingSweep(hosts)
	if len(aliveFromICMP) == 0 {
		fmt.Fprintf(&sb, "No hosts responded to ICMP/TCP alive check.\n")
	} else {
		fmt.Fprintf(&sb, "Alive hosts (%d):\n", len(aliveFromICMP))
		for _, h := range aliveFromICMP {
			fmt.Fprintf(&sb, "  %s\n", h)
		}
	}

	// === Step 2: ARP Table ===
	fmt.Fprintf(&sb, "\n=== ARP Table ===\n")
	arpRaw, arpEntries := arpTableDump()
	sb.WriteString(arpRaw)
	if len(arpEntries) == 0 {
		fmt.Fprintf(&sb, "No ARP entries found.\n")
	} else {
		fmt.Fprintf(&sb, "Parsed %d ARP entries.\n", len(arpEntries))
	}

	// === Step 3: Domain Controller Discovery ===
	fmt.Fprintf(&sb, "\n=== Domain Controller Discovery ===\n")
	sb.WriteString(nltestDCDiscovery())

	// === Step 4: DNS Reverse Lookup ===
	fmt.Fprintf(&sb, "\n=== DNS Reverse Lookup ===\n")
	// Union of ICMP alive hosts and ARP-discovered IPs
	seen := make(map[string]bool)
	var lookupTargets []string
	for _, h := range aliveFromICMP {
		if !seen[h] {
			seen[h] = true
			lookupTargets = append(lookupTargets, h)
		}
	}
	for _, e := range arpEntries {
		if !seen[e.IP] {
			seen[e.IP] = true
			lookupTargets = append(lookupTargets, e.IP)
		}
	}
	dnsResult := dnsReverseLookup(lookupTargets)
	if dnsResult == "" {
		fmt.Fprintf(&sb, "No PTR records found for alive hosts.\n")
	} else {
		sb.WriteString(dnsResult)
	}

	return NativeResult{Output: sb.String(), Success: true}, nil
}

func init() {
	Register("T1018", runT1018, nil)
}

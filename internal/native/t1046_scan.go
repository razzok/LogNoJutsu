//go:build windows

package native

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// tcpPorts is the default set of TCP ports to scan per D-04.
// Covers common service ports that SIEMs monitor for lateral movement and C2.
var tcpPorts = []int{21, 22, 23, 25, 53, 80, 135, 139, 389, 443, 445, 1433, 3306, 3389, 5985, 8080, 8443}

// udpPorts is the default set of UDP ports to scan per D-05.
// DNS(53), SNMP(161), NTP(123) — cover protocol variety for SIEM detection.
var udpPorts = []int{53, 161, 123}

// scanWorkers is the goroutine pool size — midpoint of D-08 range (10-20).
const scanWorkers = 15

// dialTimeout is the per-connection timeout for TCP/UDP dial attempts.
const dialTimeout = 300 * time.Millisecond

// scanResult holds the outcome of a single host:port scan attempt.
type scanResult struct {
	Host     string
	Port     int
	Open     bool
	Protocol string
}

// localSubnet returns the /24 CIDR of the first non-loopback IPv4 interface.
// Duplicated from engine.detectLocalSubnet() to avoid import cycle between
// internal/engine and internal/native.
// Returns "unknown" if no suitable interface is found.
func localSubnet() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ip4 := ipNet.IP.To4(); ip4 != nil {
					return fmt.Sprintf("%d.%d.%d.0/24", ip4[0], ip4[1], ip4[2])
				}
			}
		}
	}
	return "unknown"
}

// localIPAddrs collects all local IPv4 address strings from all interfaces.
// Used to exclude the host's own IPs from scan targets (D-02).
func localIPAddrs() []string {
	var result []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return result
	}
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip4 := ip.To4(); ip4 != nil {
				result = append(result, ip4.String())
			}
		}
	}
	return result
}

// subnetHosts generates the list of usable host IPs in a /24 CIDR, excluding
// the network address (.0), broadcast address (.255), and any IPs in excludeIPs.
// Returns an error if cidr is not a valid CIDR string.
func subnetHosts(cidr string, excludeIPs ...string) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("subnetHosts: invalid CIDR %q: %w", cidr, err)
	}

	// Build exclusion set
	excluded := make(map[string]bool)
	for _, ip := range excludeIPs {
		if ip != "" {
			excluded[ip] = true
		}
	}

	// Get the network base address
	base := ipNet.IP.To4()
	if base == nil {
		return nil, fmt.Errorf("subnetHosts: CIDR %q is not an IPv4 network", cidr)
	}

	var hosts []string
	// Generate .1 through .254 from the network base
	for i := 1; i <= 254; i++ {
		ip := fmt.Sprintf("%d.%d.%d.%d", base[0], base[1], base[2], i)
		if !excluded[ip] {
			hosts = append(hosts, ip)
		}
	}
	return hosts, nil
}

// scanTCP performs TCP connect scans against all hosts/ports using a bounded goroutine pool.
// The semaphore channel limits concurrency to scanWorkers simultaneous connections,
// creating a burst signature from lognojutsu.exe that matches real scanner behavior (D-08).
func scanTCP(hosts []string, ports []int) []scanResult {
	results := make(chan scanResult, len(hosts)*len(ports))
	sem := make(chan struct{}, scanWorkers)
	var wg sync.WaitGroup

	for _, host := range hosts {
		for _, port := range ports {
			wg.Add(1)
			go func(h string, p int) {
				defer wg.Done()
				sem <- struct{}{}        // acquire slot
				defer func() { <-sem }() // release slot

				addr := fmt.Sprintf("%s:%d", h, p)
				conn, err := net.DialTimeout("tcp4", addr, dialTimeout)
				if err == nil {
					conn.Close()
					results <- scanResult{Host: h, Port: p, Open: true, Protocol: "tcp"}
				}
			}(host, port)
		}
	}

	wg.Wait()
	close(results)

	var out []scanResult
	for r := range results {
		out = append(out, r)
	}
	return out
}

// scanUDP performs UDP probe scans against all hosts/ports.
// UDP is connectionless — we send a null byte and listen briefly for a response.
// Note: Windows Firewall typically suppresses ICMP port-unreachable replies, so
// most ports appear "open" in the absence of a response; results are best-effort.
func scanUDP(hosts []string, ports []int) []scanResult {
	var out []scanResult
	for _, host := range hosts {
		for _, port := range ports {
			addr := fmt.Sprintf("%s:%d", host, port)
			conn, err := net.Dial("udp4", addr)
			if err != nil {
				continue
			}
			// Send a single null byte probe
			_, _ = conn.Write([]byte{0})
			conn.SetReadDeadline(time.Now().Add(dialTimeout)) //nolint:errcheck
			buf := make([]byte, 64)
			n, _ := conn.Read(buf)
			conn.Close()
			if n > 0 {
				out = append(out, scanResult{Host: host, Port: port, Open: true, Protocol: "udp"})
			}
		}
	}
	return out
}

// runT1046 executes T1046 Network Service Discovery.
// It auto-detects the local /24 subnet, excludes own IPs, then scans all
// 17 TCP ports and 3 UDP ports using a goroutine pool.
// The burst of simultaneous connections from lognojutsu.exe PID generates
// Sysmon EID 3 events matching real port-scanner behavior (SCAN-01, SCAN-03).
func runT1046() (NativeResult, error) {
	subnet := localSubnet()
	if subnet == "unknown" {
		return NativeResult{
			Success:     false,
			ErrorOutput: "T1046: cannot detect local subnet — no active non-loopback IPv4 interface found",
		}, fmt.Errorf("T1046: cannot detect local subnet")
	}

	ownIPs := localIPAddrs()
	hosts, err := subnetHosts(subnet, ownIPs...)
	if err != nil {
		return NativeResult{
			Success:     false,
			ErrorOutput: fmt.Sprintf("T1046: subnet enumeration failed: %s", err.Error()),
		}, fmt.Errorf("T1046 subnetHosts: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Scanning %s (%d hosts, %d TCP ports, %d UDP ports)...\n",
		subnet, len(hosts), len(tcpPorts), len(udpPorts))

	// TCP connect scan — goroutine pool creates burst of simultaneous connections
	tcpResults := scanTCP(hosts, tcpPorts)

	// UDP probe scan — best-effort, Windows Firewall suppresses most responses
	udpResults := scanUDP(hosts, udpPorts)
	if len(udpResults) > 0 {
		fmt.Fprintf(&sb, "Note: UDP results are best-effort — Windows Firewall may suppress ICMP port-unreachable replies.\n")
	}

	// Group results by host
	openByHost := make(map[string][]string)
	for _, r := range tcpResults {
		if r.Open {
			openByHost[r.Host] = append(openByHost[r.Host], fmt.Sprintf("%s/%d", r.Protocol, r.Port))
		}
	}
	for _, r := range udpResults {
		if r.Open {
			openByHost[r.Host] = append(openByHost[r.Host], fmt.Sprintf("%s/%d", r.Protocol, r.Port))
		}
	}

	if len(openByHost) == 0 {
		fmt.Fprintf(&sb, "No open ports found — hosts may be down or firewalled.\n")
	} else {
		fmt.Fprintf(&sb, "Open ports found on %d host(s):\n", len(openByHost))
		for host, ports := range openByHost {
			fmt.Fprintf(&sb, "  %s: %s\n", host, strings.Join(ports, ", "))
		}
	}

	fmt.Fprintf(&sb, "Sysmon EID 3 burst from lognojutsu.exe PID — scan complete.\n")

	return NativeResult{Output: sb.String(), Success: true}, nil
}

func init() {
	Register("T1046", runT1046, nil)
}

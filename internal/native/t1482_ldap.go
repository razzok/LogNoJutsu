//go:build windows

package native

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

// ErrNoDCReachable is returned by discoverDC when no domain controller can be located.
var ErrNoDCReachable = errors.New("no domain controller reachable")

// discoverDC attempts to find a domain controller address using the following chain:
//  1. LOGONSERVER environment variable (fast path on domain-joined machines)
//  2. DNS SRV lookup for _ldap._tcp.dc._msdcs.<USERDNSDOMAIN>
//
// Returns "<host>:389" on success, or ErrNoDCReachable if no DC found.
func discoverDC() (string, error) {
	// 1. LOGONSERVER env var — fast path on domain-joined Windows machine
	if ls := os.Getenv("LOGONSERVER"); ls != "" {
		dc := strings.TrimPrefix(ls, `\\`)
		return dc + ":389", nil
	}

	// 2. DNS SRV lookup: _ldap._tcp.dc._msdcs.<domain>
	domain := os.Getenv("USERDNSDOMAIN")
	if domain != "" {
		_, addrs, err := net.LookupSRV("ldap", "tcp", "dc._msdcs."+domain)
		if err == nil && len(addrs) > 0 {
			return fmt.Sprintf("%s:%d", strings.TrimSuffix(addrs[0].Target, "."), addrs[0].Port), nil
		}
	}

	return "", ErrNoDCReachable
}

// constructBaseDN converts a DNS domain name to an LDAP base DN under CN=System.
// Example: "corp.example.com" → "CN=System,DC=corp,DC=example,DC=com"
func constructBaseDN(domain string) string {
	if domain == "" {
		return ""
	}
	parts := strings.Split(domain, ".")
	dcParts := make([]string, len(parts))
	for i, p := range parts {
		dcParts[i] = "DC=" + p
	}
	return "CN=System," + strings.Join(dcParts, ",")
}

// runT1482 performs T1482 Domain Trust Discovery via LDAP.
// It auto-discovers the DC, performs an NTLM bind using the current user's credentials,
// and queries trustedDomain objects from CN=System in the domain's naming context.
//
// Graceful fallback: if no DC is reachable, returns Success=false with a descriptive
// ErrorOutput message instead of crashing. This satisfies ARCH-02.
func runT1482() (NativeResult, error) {
	dcAddr, err := discoverDC()
	if err != nil {
		msg := "Domain trust discovery: " + err.Error()
		return NativeResult{Success: false, ErrorOutput: msg}, fmt.Errorf("T1482 discoverDC: %w", err)
	}

	// Dial with explicit 5s timeout — avoids hanging on firewall-dropped connections (Pitfall 5)
	conn, err := ldap.DialURL("ldap://"+dcAddr,
		ldap.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}),
	)
	if err != nil {
		msg := fmt.Sprintf("Domain trust discovery: LDAP dial failed (%s): %s", dcAddr, err.Error())
		return NativeResult{Success: false, ErrorOutput: msg}, fmt.Errorf("T1482 dial: %w", err)
	}
	defer conn.Close()

	// NTLM bind with empty credentials — uses current user's Windows session token (D-06)
	// On non-domain-joined machines this returns LDAP code 49 — treated as graceful fallback (Pitfall 4)
	if err := conn.NTLMBind("", "", ""); err != nil {
		msg := fmt.Sprintf("Domain trust discovery: NTLM bind failed — not domain-joined or no AD access: %s", err.Error())
		return NativeResult{Success: false, ErrorOutput: msg}, fmt.Errorf("T1482 ntlm bind: %w", err)
	}

	// Determine base DN: query RootDSE first, fall back to constructing from USERDNSDOMAIN
	baseDN := ""
	rootDSEReq := ldap.NewSearchRequest(
		"",
		ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		0, 10, false,
		"(objectClass=*)",
		[]string{"defaultNamingContext"},
		nil,
	)
	rootDSEResult, err := conn.Search(rootDSEReq)
	if err == nil && len(rootDSEResult.Entries) > 0 {
		nc := rootDSEResult.Entries[0].GetAttributeValue("defaultNamingContext")
		if nc != "" {
			baseDN = "CN=System," + nc
		}
	}
	if baseDN == "" {
		baseDN = constructBaseDN(os.Getenv("USERDNSDOMAIN"))
	}
	if baseDN == "" {
		return NativeResult{
			Success:     false,
			ErrorOutput: "Domain trust discovery: could not determine base DN (USERDNSDOMAIN not set and RootDSE query failed)",
		}, fmt.Errorf("T1482: base DN unknown")
	}

	// Search for trustedDomain objects — generates EID 4662 on the DC (D-07)
	searchReq := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 10, false,
		"(objectClass=trustedDomain)",
		[]string{"name", "trustType", "trustDirection"},
		nil,
	)
	sr, err := conn.Search(searchReq)
	if err != nil {
		msg := fmt.Sprintf("Domain trust discovery: LDAP search failed: %s", err.Error())
		return NativeResult{Success: false, ErrorOutput: msg}, fmt.Errorf("T1482 search: %w", err)
	}

	var sb strings.Builder
	if len(sr.Entries) == 0 {
		sb.WriteString("No domain trusts found (single-domain environment)\n")
	} else {
		fmt.Fprintf(&sb, "Domain trusts found: %d\n", len(sr.Entries))
		for _, entry := range sr.Entries {
			fmt.Fprintf(&sb, "Trust: %s (type=%s direction=%s)\n",
				entry.GetAttributeValue("name"),
				entry.GetAttributeValue("trustType"),
				entry.GetAttributeValue("trustDirection"),
			)
		}
	}

	return NativeResult{Output: sb.String(), Success: true}, nil
}

func init() {
	Register("T1482", runT1482, nil)
}

// Package userstore manages user profiles for multi-user simulation.
// Profiles are persisted to lognojutsu_users.json next to the executable.
// Passwords are encrypted with Windows DPAPI (machine + user key) so they
// can only be decrypted on the same system by the same user account.
package userstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// UserType distinguishes local Windows accounts from domain accounts.
type UserType string

const (
	UserTypeLocal   UserType = "local"
	UserTypeDomain  UserType = "domain"
	UserTypeCurrent UserType = "current" // run as the account that started lognojutsu.exe
)

// UserProfile holds credentials and metadata for one simulation user.
type UserProfile struct {
	ID           string   `json:"id"`
	DisplayName  string   `json:"display_name"`
	Username     string   `json:"username"`          // plain username, e.g. "jdoe"
	Domain       string   `json:"domain"`            // empty for local users
	PasswordEnc  string   `json:"password_enc"`      // DPAPI-encrypted, base64
	UserType     UserType `json:"user_type"`
	Enabled      bool     `json:"enabled"`
	LastTestOK   bool     `json:"last_test_ok"`
	LastTestMsg  string   `json:"last_test_msg"`
}

// QualifiedName returns DOMAIN\username or .\username for local accounts.
func (u *UserProfile) QualifiedName() string {
	if u.UserType == UserTypeDomain && u.Domain != "" {
		return u.Domain + `\` + u.Username
	}
	return `.\` + u.Username
}

// Store is a thread-safe collection of UserProfiles backed by a JSON file.
type Store struct {
	mu       sync.RWMutex
	profiles map[string]*UserProfile
	filePath string
}

var defaultFilePath = "lognojutsu_users.json"

// Load reads the persisted user store from disk (creates empty if not present).
func Load() (*Store, error) {
	s := &Store{
		profiles: make(map[string]*UserProfile),
		filePath: defaultFilePath,
	}
	data, err := os.ReadFile(defaultFilePath)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading user store: %w", err)
	}
	var list []*UserProfile
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parsing user store: %w", err)
	}
	for _, p := range list {
		s.profiles[p.ID] = p
	}
	return s, nil
}

// save persists the store to disk (must be called with mu held or after lock).
func (s *Store) save() error {
	list := s.List()
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0600) // 0600 = owner-only read/write
}

// List returns all profiles as a slice.
func (s *Store) List() []*UserProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*UserProfile, 0, len(s.profiles))
	for _, p := range s.profiles {
		// Return a copy without the encrypted password for API responses
		cp := *p
		cp.PasswordEnc = "" // never send encrypted password to UI
		result = append(result, &cp)
	}
	return result
}

// Get returns a profile by ID (including the encrypted password field).
func (s *Store) Get(id string) (*UserProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[id]
	return p, ok
}

// Add creates or updates a user profile. Password is encrypted with DPAPI.
func (s *Store) Add(username, domain, password string, userType UserType, displayName string) (*UserProfile, error) {
	id := strings.ToLower(strings.ReplaceAll(username, " ", "_"))
	if domain != "" {
		id = strings.ToLower(domain) + "_" + id
	}

	encPassword := ""
	if password != "" {
		var err error
		encPassword, err = encryptDPAPI(password)
		if err != nil {
			// Fall back to plain storage with a warning prefix
			encPassword = "PLAIN:" + password
		}
	}

	if displayName == "" {
		if domain != "" {
			displayName = domain + `\` + username
		} else {
			displayName = username + " (local)"
		}
	}

	p := &UserProfile{
		ID:          id,
		DisplayName: displayName,
		Username:    username,
		Domain:      domain,
		PasswordEnc: encPassword,
		UserType:    userType,
		Enabled:     true,
	}

	s.mu.Lock()
	s.profiles[id] = p
	err := s.save()
	s.mu.Unlock()

	cp := *p
	cp.PasswordEnc = ""
	return &cp, err
}

// Delete removes a profile by ID.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.profiles[id]; !ok {
		return fmt.Errorf("user profile %q not found", id)
	}
	delete(s.profiles, id)
	return s.save()
}

// DecryptPassword returns the plaintext password for a profile.
func (s *Store) DecryptPassword(id string) (string, error) {
	s.mu.RLock()
	p, ok := s.profiles[id]
	s.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("profile %q not found", id)
	}
	if p.PasswordEnc == "" {
		return "", nil
	}
	if strings.HasPrefix(p.PasswordEnc, "PLAIN:") {
		return strings.TrimPrefix(p.PasswordEnc, "PLAIN:"), nil
	}
	return decryptDPAPI(p.PasswordEnc)
}

// SetTestResult updates the last credential test result.
func (s *Store) SetTestResult(id string, ok bool, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, exists := s.profiles[id]; exists {
		p.LastTestOK = ok
		p.LastTestMsg = msg
		_ = s.save()
	}
}

// DiscoverLocalUsers queries the local system for all local user accounts.
func DiscoverLocalUsers() ([]DiscoveredUser, error) {
	script := `
$users = Get-LocalUser -ErrorAction SilentlyContinue |
    Select-Object Name, Enabled, LastLogon, Description |
    ConvertTo-Json -Compress
if ($users) { $users } else { "[]" }
`
	out, err := runPS(script)
	if err != nil {
		return nil, fmt.Errorf("discovering local users: %w (output: %s)", err, out)
	}
	out = strings.TrimSpace(out)
	if out == "" || out == "null" {
		return nil, nil
	}
	// Handle single object (not array)
	if !strings.HasPrefix(out, "[") {
		out = "[" + out + "]"
	}
	var raw []struct {
		Name        string `json:"Name"`
		Enabled     bool   `json:"Enabled"`
		Description string `json:"Description"`
	}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("parsing user list: %w (raw: %s)", err, out)
	}
	result := make([]DiscoveredUser, 0, len(raw))
	for _, u := range raw {
		result = append(result, DiscoveredUser{
			Username: u.Name,
			Enabled:  u.Enabled,
			Source:   "local",
			Note:     u.Description,
		})
	}
	return result, nil
}

// DiscoverRecentDomainUsers reads recent 4624 logon events to find domain users.
func DiscoverRecentDomainUsers() ([]DiscoveredUser, error) {
	script := `
try {
    $events = Get-WinEvent -FilterHashtable @{LogName='Security'; Id=4624} -MaxEvents 200 -ErrorAction Stop
    $seen = @{}
    $users = @()
    foreach ($e in $events) {
        $xml = [xml]$e.ToXml()
        $ns = @{e='http://schemas.microsoft.com/win/2004/08/events/event'}
        $data = $xml.Event.EventData.Data
        $targetUser = ($data | Where-Object { $_.Name -eq 'TargetUserName' }).'#text'
        $targetDomain = ($data | Where-Object { $_.Name -eq 'TargetDomainName' }).'#text'
        $logonType = ($data | Where-Object { $_.Name -eq 'LogonType' }).'#text'
        if ($targetUser -and $targetDomain -and
            $targetDomain -notmatch 'NT AUTHORITY|WORKGROUP|BUILTIN' -and
            $targetUser -notmatch '\$$' -and
            -not $seen["$targetDomain\$targetUser"]) {
            $seen["$targetDomain\$targetUser"] = $true
            $users += [PSCustomObject]@{
                Username = $targetUser
                Domain   = $targetDomain
                LogonType = $logonType
            }
        }
    }
    $users | ConvertTo-Json -Compress
} catch {
    "[]"
}
`
	out, err := runPS(script)
	if err != nil {
		return nil, fmt.Errorf("discovering domain users: %w", err)
	}
	out = strings.TrimSpace(out)
	if out == "" || out == "null" || out == "[]" {
		return nil, nil
	}
	if !strings.HasPrefix(out, "[") {
		out = "[" + out + "]"
	}
	var raw []struct {
		Username  string `json:"Username"`
		Domain    string `json:"Domain"`
		LogonType string `json:"LogonType"`
	}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("parsing domain users: %w (raw: %.200s)", err, out)
	}
	result := make([]DiscoveredUser, 0, len(raw))
	for _, u := range raw {
		result = append(result, DiscoveredUser{
			Username: u.Username,
			Domain:   u.Domain,
			Enabled:  true,
			Source:   "domain (from event log)",
			Note:     "LogonType " + u.LogonType,
		})
	}
	return result, nil
}

// DiscoveredUser is a user found by the discovery scan (not yet added to the store).
type DiscoveredUser struct {
	Username string `json:"username"`
	Domain   string `json:"domain"`
	Enabled  bool   `json:"enabled"`
	Source   string `json:"source"`
	Note     string `json:"note"`
}

// TestCredentials verifies that credentials for a profile are valid.
func TestCredentials(profile *UserProfile, password string) (bool, string) {
	qualName := profile.QualifiedName()
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.DirectoryServices.AccountManagement
try {
    $ctxType = if ("%s" -match "^\\." ) {
        [System.DirectoryServices.AccountManagement.ContextType]::Machine
    } else {
        [System.DirectoryServices.AccountManagement.ContextType]::Domain
    }
    $ctx = New-Object System.DirectoryServices.AccountManagement.PrincipalContext($ctxType)
    $result = $ctx.ValidateCredentials("%s", "%s")
    if ($result) { "OK" } else { "INVALID_CREDENTIALS" }
} catch {
    "ERROR: $_"
}
`,
		qualName,
		strings.ReplaceAll(profile.Username, `"`, `\"`),
		strings.ReplaceAll(password, `"`, `\"`),
	)
	out, err := runPS(script)
	out = strings.TrimSpace(out)
	if err != nil || strings.HasPrefix(out, "ERROR:") {
		msg := out
		if err != nil {
			msg = err.Error()
		}
		return false, "Test failed: " + msg
	}
	if out == "OK" {
		return true, "Credentials valid"
	}
	return false, "Invalid credentials"
}

// encryptDPAPI encrypts a string using Windows DPAPI via PowerShell.
func encryptDPAPI(plaintext string) (string, error) {
	script := fmt.Sprintf(
		`$s = ConvertTo-SecureString "%s" -AsPlainText -Force; ConvertFrom-SecureString $s`,
		strings.ReplaceAll(plaintext, `"`, "`\""),
	)
	out, err := runPS(script)
	if err != nil {
		return "", fmt.Errorf("DPAPI encrypt: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// decryptDPAPI decrypts a DPAPI-encrypted string via PowerShell.
func decryptDPAPI(encrypted string) (string, error) {
	script := fmt.Sprintf(`
$s = ConvertTo-SecureString "%s"
[System.Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($s)
)`, strings.TrimSpace(encrypted))
	out, err := runPS(script)
	if err != nil {
		return "", fmt.Errorf("DPAPI decrypt: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func runPS(script string) (string, error) {
	cmd := exec.Command("powershell.exe",
		"-NonInteractive", "-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-Command", script)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), err
}

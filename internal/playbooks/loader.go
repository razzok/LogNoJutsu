package playbooks

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed embedded
var embeddedFS embed.FS

// Registry holds all loaded techniques and campaigns.
type Registry struct {
	Techniques map[string]*Technique
	Campaigns  map[string]*Campaign
}

// LoadEmbedded loads all playbooks bundled into the binary.
func LoadEmbedded() (*Registry, error) {
	r := &Registry{
		Techniques: make(map[string]*Technique),
		Campaigns:  make(map[string]*Campaign),
	}

	// Load techniques
	err := fs.WalkDir(embeddedFS, "embedded/techniques", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}
		data, readErr := embeddedFS.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("reading %s: %w", path, readErr)
		}
		var t Technique
		if parseErr := yaml.Unmarshal(data, &t); parseErr != nil {
			return fmt.Errorf("parsing %s: %w", path, parseErr)
		}
		r.Techniques[t.ID] = &t
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading techniques: %w", err)
	}

	// Load campaigns
	err = fs.WalkDir(embeddedFS, "embedded/campaigns", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}
		data, readErr := embeddedFS.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("reading %s: %w", path, readErr)
		}
		var c Campaign
		if parseErr := yaml.Unmarshal(data, &c); parseErr != nil {
			return fmt.Errorf("parsing %s: %w", path, parseErr)
		}
		r.Campaigns[c.ID] = &c
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading campaigns: %w", err)
	}

	return r, nil
}

// GetTechniquesByPhase returns all techniques for a given phase.
func (r *Registry) GetTechniquesByPhase(phase string) []*Technique {
	var result []*Technique
	for _, t := range r.Techniques {
		if t.Phase == phase {
			result = append(result, t)
		}
	}
	return result
}

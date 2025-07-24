package data

import (
	"math"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// ArchitectYAML represents the structure of architect.yaml
type ArchitectYAML struct {
	ProjectMeta   map[string]string `yaml:"project_meta,omitempty"`
	BooleanFields []struct {
		Name    string `yaml:"name"`
		Address int    `yaml:"address"`
		Bit     *int   `yaml:"bit,omitempty"`
	} `yaml:"boolean_fields"`
	FaultFields []struct {
		Name    string `yaml:"name"`
		Address int    `yaml:"address"`
		Bit     *int   `yaml:"bit,omitempty"`
	} `yaml:"fault_fields"`
	// FloatFields are grouped by subgroup (e.g., "Performance", "HopperVibratory")
	FloatFields map[string][]struct {
		Name    string `yaml:"name"`
		Address int    `yaml:"address"`
	} `yaml:"float_fields"`
}

// ParsePLCDataFromRegisters uses the cached architect.yaml mapping to parse raw register data.
// This function is optimized to avoid file I/O on every call by using the in-memory cache.
func ParsePLCDataFromRegisters(registers []uint16) (map[string]interface{}, error) {
	arch := GetArchitectYAML() // Use the cached version

	result := make(map[string]interface{})

	// The ProjectMeta field from `arch` is intentionally ignored here,
	// as this function is only concerned with parsing PLC register data.
	// Booleans
	for _, field := range arch.BooleanFields {
		reg := registers[field.Address]
		bit := 0
		if field.Bit != nil {
			bit = *field.Bit
		}
		val := (reg & (1 << bit)) != 0
		result[field.Name] = val
	}

	// Faults
	for _, field := range arch.FaultFields {
		reg := registers[field.Address]
		bit := 0
		if field.Bit != nil {
			bit = *field.Bit
		}
		val := (reg & (1 << bit)) != 0
		result[field.Name] = val
	}

	// Floats (robust pairing of HighINT/LowINT within each group)
	for groupName, fields := range arch.FloatFields {
		floatPairs := make(map[string]struct{ High, Low *int })
		for i := 0; i < len(fields); i++ {
			name := fields[i].Name
			addr := fields[i].Address
			if strings.HasSuffix(name, "(HighINT)") {
				base := strings.TrimSuffix(name, "(HighINT)")
				p := floatPairs[base]
				p.High = &addr
				floatPairs[base] = p
			} else if strings.HasSuffix(name, "(LowINT)") {
				base := strings.TrimSuffix(name, "(LowINT)")
				p := floatPairs[base]
				p.Low = &addr
				floatPairs[base] = p
			}
		}
		for base, p := range floatPairs {
			if p.High != nil && p.Low != nil {
				high := uint32(registers[*p.High])
				low := uint32(registers[*p.Low])
				f := math.Float32frombits((high << 16) | low)
				// Create a namespaced field name for InfluxDB, e.g., "Floats.Performance.PartsPerMinute"
				namespacedKey := "Floats." + groupName + "." + base
				result[namespacedKey] = f
			}
		}
	}

	return result, nil
}

// Package-level cache for ArchitectYAML
var CachedArchitectYAML *ArchitectYAML

// GetArchitectYAML returns the cached ArchitectYAML, or logs fatal if not loaded
func GetArchitectYAML() *ArchitectYAML {
	if CachedArchitectYAML == nil {
		panic("ArchitectYAML not loaded. Call LoadAndCacheArchitectYAML at startup.")
	}
	return CachedArchitectYAML
}

// GetProjectMeta returns the project metadata from the cached ArchitectYAML.
// It returns nil if the cache is not loaded or if project_meta is not present.
func GetProjectMeta() map[string]string {
	if CachedArchitectYAML == nil {
		return nil
	}
	return CachedArchitectYAML.ProjectMeta
}

// LoadAndCacheArchitectYAML loads architect.yaml and caches it in memory
func LoadAndCacheArchitectYAML(path string) error {
	arch, err := LoadArchitectYAMLFromPath(path)
	if err != nil {
		return err
	}
	CachedArchitectYAML = arch
	return nil
}

// LoadArchitectYAMLFromPath loads and parses architect.yaml from a given path
func LoadArchitectYAMLFromPath(path string) (*ArchitectYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var arch ArchitectYAML
	err = yaml.Unmarshal(data, &arch)
	if err != nil {
		return nil, err
	}
	return &arch, nil
}

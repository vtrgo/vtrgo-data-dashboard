// file: service/data/architect.go
//
//	Data structures and functions for parsing PLC data from registers
package data

import (
	"fmt"
	"math"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// ArchitectYAML represents the structure of the architect.yaml configuration file,
// which defines how raw PLC data is mapped to meaningful fields.
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

// ParsePLCDataFromRegisters uses the cached architect.yaml mapping to parse raw
// register data into a map of field names to their corresponding values.
// It is optimized to avoid file I/O on every call by using an in-memory cache.
func ParsePLCDataFromRegisters(registers []uint16) (map[string]interface{}, error) {
	arch, err := GetArchitectYAML()
	if err != nil {
		return nil, err
	}

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

// CachedArchitectYAML holds the in-memory representation of the architect.yaml file.
// It is populated at startup by LoadAndCacheArchitectYAML.
var CachedArchitectYAML *ArchitectYAML

// GetArchitectYAML returns the cached ArchitectYAML configuration, returning an
// error if it has not been initialized by calling LoadAndCacheArchitectYAML.
func GetArchitectYAML() (*ArchitectYAML, error) {
	if CachedArchitectYAML == nil {
		return nil, fmt.Errorf("ArchitectYAML not loaded. Call LoadAndCacheArchitectYAML at startup")
	}
	return CachedArchitectYAML, nil
}

// GetProjectMeta returns the project metadata from the cached ArchitectYAML.
// It returns an error if the cache is not loaded.
func GetProjectMeta() (map[string]string, error) {
	arch, err := GetArchitectYAML()
	if err != nil {
		return nil, err
	}
	return arch.ProjectMeta, nil
}

// LoadAndCacheArchitectYAML reads the architect.yaml file from the given path,
// parses it, and stores it in a package-level cache for fast access.
func LoadAndCacheArchitectYAML(path string) error {
	arch, err := LoadArchitectYAMLFromPath(path)
	if err != nil {
		return err
	}
	CachedArchitectYAML = arch
	return nil
}

// LoadArchitectYAMLFromPath loads and parses the architect.yaml file from a given path.
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

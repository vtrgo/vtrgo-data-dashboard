package data

import (
	"math"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// ArchitectYAML represents the structure of architect.yaml
type ArchitectYAML struct {
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
	FloatFields []struct {
		Name    string `yaml:"name"`
		Address int    `yaml:"address"`
	} `yaml:"float_fields"`
}

// LoadPLCDataMapFromYAML loads the PLCDataMap from registers using architect.yaml mapping (generic for booleans, faults, floats)
func LoadPLCDataMapFromYAML(yamlPath string, registers []uint16) (map[string]interface{}, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}
	var arch ArchitectYAML
	err = yaml.Unmarshal(data, &arch)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

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

	// Floats (robust pairing of HighINT/LowINT)
	floatPairs := make(map[string]struct{ High, Low *int })
	for i := 0; i < len(arch.FloatFields); i++ {
		name := arch.FloatFields[i].Name
		addr := arch.FloatFields[i].Address
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
			result[base] = f
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

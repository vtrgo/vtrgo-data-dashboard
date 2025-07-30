// file: service/data/converter.go
package data

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// NOTE: These structs define the structure of the architect.yaml file.
// They are likely defined elsewhere in the `data` package but are included
// here for clarity. They should be the same structs used by
// `LoadAndCacheArchitectYAML` and `GetArchitectYAML`.

type PLCFieldYAML struct {
	Name    string `yaml:"name"`
	Address int    `yaml:"address"`
	Bit     *int   `yaml:"bit,omitempty"`
}

type FloatFieldYAML struct {
	Name    string `yaml:"name"`
	Address int    `yaml:"address"`
}

type PLCDataMapYAML struct {
	ProjectMeta   map[string]string           `yaml:"project_meta,omitempty"`
	BooleanFields []PLCFieldYAML              `yaml:"boolean_fields"`
	FaultFields   []PLCFieldYAML              `yaml:"fault_fields"`
	FloatFields   map[string][]FloatFieldYAML `yaml:"float_fields"`
}

// parseSpecifier parses e.g. ModbusDataWrite[1].10 into address=1, bit=10
func parseSpecifier(spec string) (address int, bit int, err error) {
	re := regexp.MustCompile(`ModbusDataWrite\[(\d+)\](?:\.(\d+))?`)
	matches := re.FindStringSubmatch(spec)
	if len(matches) >= 2 {
		address, _ = strconv.Atoi(matches[1])
	}
	if len(matches) == 3 && matches[2] != "" {
		bit, _ = strconv.Atoi(matches[2])
	}
	return
}

// isBooleanGroup checks if the description belongs to a known boolean group.
// This makes the parser explicit about what it handles, preventing miscategorization.
func isBooleanGroup(desc string) bool {
	knownBooleanPrefixes := []string{
		"SystemStatusBits",
		"FeederStatusBits",
		"RobotStatusBits",
	}
	for _, prefix := range knownBooleanPrefixes {
		if strings.HasPrefix(desc, prefix) {
			return true
		}
	}
	return false
}

// CSVToYAML converts CSV data from an io.Reader into a structured YAML file.
// It parses PLC data mappings from the CSV and writes them to the specified yamlPath.
func CSVToYAML(csvInput io.Reader, yamlPath string) error {
	reader := csv.NewReader(csvInput)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	out := PLCDataMapYAML{
		ProjectMeta: make(map[string]string),
	}
	// Use a temporary map to collect floats by their sub-group.
	out.FloatFields = make(map[string][]FloatFieldYAML)
	// Find the header row index, and process remarks along the way
	headerIndex := -1
	for i, row := range records {
		if len(row) > 0 && row[0] == "remark" {
			if len(row) >= 3 {
				key := strings.TrimSpace(row[1])
				value := strings.TrimSpace(row[2])
				if key != "" {
					out.ProjectMeta[key] = value
				}
			}
			continue // go to next row
		}

		if len(row) >= 4 && row[0] == "TYPE" && row[1] == "SCOPE" && row[2] == "NAME" && row[3] == "DESCRIPTION" {
			headerIndex = i
			break
		}
	}

	// If no project meta was found, make the map nil so it's omitted from YAML
	if len(out.ProjectMeta) == 0 {
		out.ProjectMeta = nil
	}

	if headerIndex == -1 {
		return fmt.Errorf("header row (starting with TYPE,SCOPE,NAME,DESCRIPTION) not found in CSV file")
	}

	for _, row := range records[headerIndex+1:] { // Start processing from the line after the header
		if len(row) < 6 {
			continue
		}
		desc := row[3]
		spec := row[5]
		if spec == "" {
			continue
		}
		address, bit, _ := parseSpecifier(spec)

		var bitPtr *int
		if strings.Contains(spec, ".") { // Only set pointer if bit is present
			bitPtr = new(int)
			*bitPtr = bit
		}

		trimmedDesc := strings.TrimSpace(desc)
		parts := strings.Split(trimmedDesc, " - ")
		mainGroup := parts[0]

		if strings.HasPrefix(trimmedDesc, "FaultBits") || strings.HasPrefix(trimmedDesc, "WarningBits") {
			nameNorm := strings.ReplaceAll(strings.Join(parts, "."), " ", "")
			out.FaultFields = append(out.FaultFields, PLCFieldYAML{Name: nameNorm, Address: address, Bit: bitPtr})
		} else if strings.HasPrefix(trimmedDesc, "Floats") {
			re := regexp.MustCompile(`^Floats\s*-\s*(.*?)\s*-\s*(.*)$`)
			matches := re.FindStringSubmatch(trimmedDesc)
			if len(matches) != 3 {
				return fmt.Errorf("malformed float description: '%s'. Expected format 'Floats - SubGroup - FieldName'", desc)
			}
			subGroup := strings.ReplaceAll(matches[1], " ", "")
			fieldName := strings.ReplaceAll(matches[2], " ", "")
			out.FloatFields[subGroup] = append(out.FloatFields[subGroup], FloatFieldYAML{Name: fieldName, Address: address})
		} else if isBooleanGroup(trimmedDesc) {
			nameNorm := strings.ReplaceAll(strings.Join(parts, "."), " ", "")
			out.BooleanFields = append(out.BooleanFields, PLCFieldYAML{Name: nameNorm, Address: address, Bit: bitPtr})
		} else {
			return fmt.Errorf("unrecognized group '%s' in CSV description: '%s'. Please update the converter to handle this group", mainGroup, desc)
		}
	}
	// Sort fields within each float group by address for deterministic output.
	for _, fields := range out.FloatFields {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Address < fields[j].Address })
	}
	outBytes, err := yaml.Marshal(out)
	if err != nil {
		return err
	}
	return os.WriteFile(yamlPath, outBytes, 0644)
}

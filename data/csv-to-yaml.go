package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type PLCFieldYAML struct {
	Name    string `yaml:"name"`
	Address int    `yaml:"address"`
	Bit     *int   `yaml:"bit,omitempty"`
}

type VibrationGroup struct {
	Name   string         `yaml:"name"`
	Fields []PLCFieldYAML `yaml:"fields"`
}

type FloatFieldYAML struct {
	Name    string `yaml:"name"`
	Address int    `yaml:"address"`
}

type FloatGroup struct {
	Name   string           `yaml:"name"`
	Fields []FloatFieldYAML `yaml:"fields"`
}

type PLCDataMapYAML struct {
	BooleanFields []PLCFieldYAML `yaml:"boolean_fields"`
	FaultFields   []PLCFieldYAML `yaml:"fault_fields"`
	FloatFields   []FloatGroup   `yaml:"float_fields"`
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

func extractGroup(fieldName string) string {
	// Use the prefix before the first " - " as the group
	if idx := strings.Index(fieldName, " - "); idx != -1 {
		return fieldName[:idx]
	}
	return fieldName
}

func CSVToYAML(csvPath, yamlPath string) error {
	f, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	var out PLCDataMapYAML
	floatGroups := make(map[string]*FloatGroup)

	for _, row := range records[2:] { // skip header lines
		if len(row) < 6 {
			continue
		}
		desc := row[3]
		spec := row[5]
		if spec == "" {
			continue
		}
		address, bit, _ := parseSpecifier(spec)
		fieldName := desc
		if fieldName == "" {
			fieldName = row[2]
		}

		var bitPtr *int
		if spec != "" {
			if strings.Contains(spec, ".") { // Only set pointer if bit is present
				bitPtr = new(int)
				*bitPtr = bit
			}
		}
		// Influx-friendly field name: replace ' - ' with '.' and remove spaces
		influxFieldName := strings.ReplaceAll(fieldName, " - ", ".")
		influxFieldName = strings.ReplaceAll(influxFieldName, " ", "")

		// Improved generic float group detection: handle Floats- and Floats. prefixes
		floatGroupMatch := false
		var groupName, fieldBase string
		if strings.HasPrefix(influxFieldName, "Floats-") {
			// e.g. Floats-VibrationData[0].VibrationX(HighINT)
			rest := strings.TrimPrefix(influxFieldName, "Floats-")
			parts := strings.SplitN(rest, ".", 2)
			if len(parts) == 2 {
				groupName = parts[0] // e.g. VibrationData[0]
				fieldBase = "Floats." + groupName + "." + parts[1]
				floatGroupMatch = true
			}
		} else if strings.HasPrefix(influxFieldName, "Floats.") {
			// e.g. Floats.VibrationData[0].VibrationX
			parts := strings.SplitN(influxFieldName, ".", 3)
			if len(parts) == 3 {
				groupName = parts[1]
				fieldBase = influxFieldName
				floatGroupMatch = true
			}
		}
		if floatGroupMatch {
			field := FloatFieldYAML{Name: fieldBase, Address: address}
			if _, ok := floatGroups[groupName]; !ok {
				floatGroups[groupName] = &FloatGroup{Name: groupName}
			}
			floatGroups[groupName].Fields = append(floatGroups[groupName].Fields, field)
			continue
		}

		group := extractGroup(fieldName)
		if group == "FaultBits" {
			out.FaultFields = append(out.FaultFields, PLCFieldYAML{
				Name:    influxFieldName,
				Address: address,
				Bit:     bitPtr,
			})
		} else if group != "Floats" {
			out.BooleanFields = append(out.BooleanFields, PLCFieldYAML{
				Name:    influxFieldName,
				Address: address,
				Bit:     bitPtr,
			})
		}
	}
	// Add float groups to output
	for _, g := range floatGroups {
		out.FloatFields = append(out.FloatFields, *g)
	}

	outBytes, err := yaml.Marshal(out)
	if err != nil {
		return err
	}
	return os.WriteFile(yamlPath, outBytes, 0644)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: csv-to-yaml <input.csv> <output.yaml>")
		os.Exit(1)
	}
	err := CSVToYAML(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println("YAML written to", os.Args[2])
}

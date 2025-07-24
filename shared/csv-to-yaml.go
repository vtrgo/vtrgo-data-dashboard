package main

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
	ProjectMeta   map[string]string `yaml:"project_meta,omitempty"`
	BooleanFields []PLCFieldYAML    `yaml:"boolean_fields"`
	FaultFields   []PLCFieldYAML    `yaml:"fault_fields"`
	FloatFields   []interface{}     `yaml:"float_fields"`
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

	out := PLCDataMapYAML{
		ProjectMeta: make(map[string]string),
	}
	type floatFieldSimple struct {
		Name    string `yaml:"name"`
		Address int    `yaml:"address"`
	}
	var floatFields []floatFieldSimple

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

		// Normalize field name: replace ' - ' with '.' and remove all spaces
		nameNorm := strings.ReplaceAll(fieldName, " - ", ".")
		nameNorm = strings.ReplaceAll(nameNorm, " ", "")

		// Improved float field detection and naming
		floatRe := regexp.MustCompile(`(?i)^float[s]?\s*[-.]+\s*`)
		if floatRe.MatchString(fieldName) {
			floatName := floatRe.ReplaceAllString(fieldName, "Floats.")
			floatName = strings.ReplaceAll(floatName, " - ", ".")
			floatName = strings.ReplaceAll(floatName, " ", "")
			floatFields = append(floatFields, floatFieldSimple{
				Name:    floatName,
				Address: address,
			})
			continue
		}

		group := extractGroup(fieldName)
		if group == "FaultBits" {
			out.FaultFields = append(out.FaultFields, PLCFieldYAML{
				Name:    nameNorm,
				Address: address,
				Bit:     bitPtr,
			})
		} else if group != "Floats" {
			out.BooleanFields = append(out.BooleanFields, PLCFieldYAML{
				Name:    nameNorm,
				Address: address,
				Bit:     bitPtr,
			})
		}
	}
	// Sort float fields by address before appending
	if len(floatFields) > 1 {
		for i := 1; i < len(floatFields); i++ {
			j := i
			for j > 0 && floatFields[j-1].Address > floatFields[j].Address {
				floatFields[j-1], floatFields[j] = floatFields[j], floatFields[j-1]
				j--
			}
		}
	}
	for _, ff := range floatFields {
		out.FloatFields = append(out.FloatFields, ff)
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

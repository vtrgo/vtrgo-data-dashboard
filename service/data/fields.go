// file: service/data/fields.go
// Contains utility functions to retrieve field names from the cached
// architect.yaml configuration. These functions are used to construct InfluxDB
// queries.
package data

import "regexp"

// GetBooleanFieldNames retrieves the list of boolean field names from the cached
// architect.yaml configuration. These names are used for constructing InfluxDB
// queries.
func GetBooleanFieldNames() ([]string, error) {
	mapping, err := GetArchitectYAML()
	if err != nil {
		return nil, err
	}
	fields := make([]string, 0, len(mapping.BooleanFields))
	for _, f := range mapping.BooleanFields {
		fields = append(fields, f.Name)
	}
	return fields, nil
}

// GetFaultFieldNames retrieves the list of fault field names from the cached
// architect.yaml configuration. These names are used for constructing InfluxDB
// queries.
func GetFaultFieldNames() ([]string, error) {
	mapping, err := GetArchitectYAML()
	if err != nil {
		return nil, err
	}
	fields := make([]string, 0, len(mapping.FaultFields))
	for _, f := range mapping.FaultFields {
		fields = append(fields, f.Name)
	}
	return fields, nil
}

// GetCombinedFloatFields generates the namespaced float field names for InfluxDB queries.
// It combines group names with field names (e.g., "Performance.PartsPerMinute").
func GetCombinedFloatFields(arch *ArchitectYAML) []string {
	re := regexp.MustCompile(`\([^)]+\)`)
	var result []string
	// Iterate over groups to build namespaced field names
	for groupName, fields := range arch.FloatFields {
		uniqueBaseNames := make(map[string]struct{})
		for _, f := range fields {
			baseName := re.ReplaceAllString(f.Name, "")
			if _, exists := uniqueBaseNames[baseName]; !exists {
				uniqueBaseNames[baseName] = struct{}{}
				result = append(result, "Floats."+groupName+"."+baseName)
			}
		}
	}
	return result
}

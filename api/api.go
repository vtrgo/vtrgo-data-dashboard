package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"vtarchitect/config"
	"vtarchitect/data"
	"vtarchitect/influx"
)

// --- YAML-driven field helpers ---

// GetBooleanFieldNames loads boolean field names from cached architect.yaml
func GetBooleanFieldNames() ([]string, error) {
	mapping := data.GetArchitectYAML()
	fields := make([]string, 0, len(mapping.BooleanFields))
	for _, f := range mapping.BooleanFields {
		fields = append(fields, f.Name)
	}
	return fields, nil
}

// GetFaultFieldNames loads fault field names from cached architect.yaml
func GetFaultFieldNames() ([]string, error) {
	mapping := data.GetArchitectYAML()
	fields := make([]string, 0, len(mapping.FaultFields))
	for _, f := range mapping.FaultFields {
		fields = append(fields, f.Name)
	}
	return fields, nil
}

// GetFloatFieldNames loads all float field names from cached architect.yaml (flattened)
func GetFloatFieldNames() ([]string, error) {
	mapping := data.GetArchitectYAML()
	fields := make([]string, 0, len(mapping.FloatFields))
	for _, f := range mapping.FloatFields {
		fields = append(fields, f.Name)
	}
	return fields, nil
}

// ---

func isValidFluxTime(input string) bool {
	if input == "now()" || (len(input) > 1 && input[0] == '-') {
		return true
	}
	_, err := time.Parse(time.RFC3339, input)
	return err == nil
}

// Helper to strip (HighINT)/(LowINT) and deduplicate
func getCombinedFloatFields(floatFields []struct {
	Name    string `yaml:"name"`
	Address int    `yaml:"address"`
}) []string {
	re := regexp.MustCompile(`\([^)]+\)`)
	unique := make(map[string]struct{})
	result := make([]string, 0, len(floatFields))
	for _, f := range floatFields {
		combined := re.ReplaceAllString(f.Name, "")
		if _, exists := unique[combined]; !exists {
			unique[combined] = struct{}{}
			result = append(result, combined)
		}
	}
	return result
}

func generateBooleanPercentagesFluxQueryFile(csvPath string, yamlCache *data.ArchitectYAML) error {
	// Compose regex for _field filter from boolean field names
	fields := make([]string, 0, len(yamlCache.BooleanFields))
	for _, f := range yamlCache.BooleanFields {
		fields = append(fields, f.Name)
	}
	// Compose regex: ^Field1|^Field2|^Field3
	regex := "^" + strings.Join(fields, "|^")

	flux := `from(bucket: "vtrFeederData")
  |> range(start: v.timeRangeStart, stop: v.timeRangeStop)
  |> filter(fn: (r) =>
    r._measurement == "status_data" and
    r._field =~ /` + regex + `/
  )
  |> keep(columns: ["_time", "_field", "_value"])
  |> map(fn: (r) => ({ r with _value: if bool(v: r._value) then 1.0 else 0.0 }))
  |> group(columns: ["_field"])
  |> mean()
  |> map(fn: (r) => ({ r with _value: r._value * 100.0 }))
  |> rename(columns: {_value: "boolean_percentage"})
`
	examplesDir := filepath.Join(".", "examples")
	os.MkdirAll(examplesDir, 0755)
	return os.WriteFile(filepath.Join(examplesDir, "flux-query-boolean-percentages.iql"), []byte(flux), 0644)
}

func StartAPIServer(cfg *config.Config, client *influx.Client) {
	// --- CSV/Flux query file generation on startup ---
	csvPath := "./your-csv-file.csv" // Change this to your actual CSV path if needed
	if _, err := os.Stat(csvPath); err == nil {
		yamlCache := data.GetArchitectYAML()
		if err := generateBooleanPercentagesFluxQueryFile(csvPath, yamlCache); err != nil {
			log.Printf("Failed to generate flux query file: %v", err)
		} else {
			log.Printf("Generated ./examples/flux-query-boolean-percentages.iql")
		}
	}

	http.HandleFunc("/api/percentages", func(w http.ResponseWriter, r *http.Request) {
		bucket := r.URL.Query().Get("bucket")
		if bucket == "" {
			bucket = cfg.Values["INFLUXDB_BUCKET"]
		}

		start := r.URL.Query().Get("start")
		if start == "" {
			start = "-1h"
		} else if !isValidFluxTime(start) {
			http.Error(w, "Invalid start time format", http.StatusBadRequest)
			return
		}

		stop := r.URL.Query().Get("stop")
		if stop == "" {
			stop = "now()"
		} else if !isValidFluxTime(stop) {
			http.Error(w, "Invalid stop time format", http.StatusBadRequest)
			return
		}

		fields, err := GetBooleanFieldNames()
		if err != nil {
			http.Error(w, "Failed to load boolean field names", http.StatusInternalServerError)
			return
		}
		measurement := cfg.Values["INFLUXDB_MEASUREMENT"]
		results, err := client.AggregateBooleanPercentages(measurement, bucket, fields, start, stop)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		bucket := r.URL.Query().Get("bucket")
		if bucket == "" {
			bucket = cfg.Values["INFLUXDB_BUCKET"]
		}

		start := r.URL.Query().Get("start")
		if start == "" {
			start = "-1h"
		} else if !isValidFluxTime(start) {
			http.Error(w, "Invalid start time format", http.StatusBadRequest)
			return
		}

		stop := r.URL.Query().Get("stop")
		if stop == "" {
			stop = "now()"
		} else if !isValidFluxTime(stop) {
			http.Error(w, "Invalid stop time format", http.StatusBadRequest)
			return
		}

		// Load field lists from YAML (use cached)
		arch := data.GetArchitectYAML()

		// Always use the Influx field/tag names from YAML cache for queries
		booleanFields := make([]string, 0, len(arch.BooleanFields))
		for _, f := range arch.BooleanFields {
			booleanFields = append(booleanFields, f.Name)
		}
		faultFields := make([]string, 0, len(arch.FaultFields))
		for _, f := range arch.FaultFields {
			faultFields = append(faultFields, f.Name)
		}
		// Use only the combined/actual influx float fields
		floatFields := getCombinedFloatFields(arch.FloatFields)

		measurement := cfg.Values["INFLUXDB_MEASUREMENT"]
		if measurement == "" {
			measurement = "status_data"
		}

		// Debug: log the actual field names used for queries
		log.Printf("Boolean fields (query): %+v", booleanFields)
		log.Printf("Fault fields (query): %+v", faultFields)
		log.Printf("Float fields (query): %+v", floatFields)
		log.Printf("Measurement: %s, Bucket: %s, Start: %s, Stop: %s", measurement, bucket, start, stop)

		// Aggregate booleans (percentage true)
		boolResults, err := client.AggregateBooleanPercentages(measurement, bucket, booleanFields, start, stop)
		if err != nil {
			http.Error(w, "Boolean aggregation error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Aggregate faults (count true)
		faultResults, err := client.AggregateFaultCounts(measurement, bucket, faultFields, start, stop)
		if err != nil {
			http.Error(w, "Fault aggregation error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Aggregate floats (mean)
		floatResults, err := client.AggregateFloatMeans(measurement, bucket, floatFields, start, stop)
		if err != nil {
			http.Error(w, "Float aggregation error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Float aggregation results: %+v", floatResults)

		results := map[string]interface{}{
			"boolean_percentages": boolResults,
			"fault_counts":        faultResults,
			"float_averages":      floatResults,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	log.Println("API server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

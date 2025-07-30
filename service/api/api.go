// file: service/api/api.go
// API server for VTArchitect that serves various endpoints
package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"vtarchitect/config"
	"vtarchitect/data"
	"vtarchitect/influx"
)

//go:embed static/* static/**/*
var staticFiles embed.FS

// StatsResponse defines the structure for the /api/stats endpoint response.
type StatsResponse struct {
	ProjectMeta        map[string]string  `json:"project_meta,omitempty"`
	BooleanPercentages map[string]float64 `json:"boolean_percentages"`
	FaultCounts        map[string]float64 `json:"fault_counts"`
	FloatAverages      map[string]float64 `json:"float_averages"`
}

// ---

func isValidFluxTime(input string) bool {
	if input == "now()" || (len(input) > 1 && input[0] == '-') {
		return true
	}
	_, err := time.Parse(time.RFC3339, input)
	return err == nil
}

// parseTimeRange extracts and validates 'start' and 'stop' query parameters from
// an HTTP request. It provides default values ("-1h" for start, "now()" for stop)
// if they are not present. It returns an error if the time formats are invalid.
func parseTimeRange(r *http.Request) (start, stop string, err error) {
	start = r.URL.Query().Get("start")
	if start == "" {
		start = "-1h" // Default start time
	} else if !isValidFluxTime(start) {
		return "", "", fmt.Errorf("invalid start time format")
	}

	stop = r.URL.Query().Get("stop")
	if stop == "" {
		stop = "now()" // Default stop time
	} else if !isValidFluxTime(stop) {
		return "", "", fmt.Errorf("invalid stop time format")
	}
	return start, stop, nil
}

// respondWithError is a helper to send a JSON error message with a status code.
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// StartAPIServer initializes and starts the HTTP server. It sets up all API
// handlers for querying data and uploading configurations, and also serves the
// static frontend application. This function blocks and should typically be run
// in a separate goroutine.
func StartAPIServer(cfg *config.Config, client *influx.Client) {
	http.HandleFunc("/api/percentages", func(w http.ResponseWriter, r *http.Request) {
		bucket := r.URL.Query().Get("bucket")
		if bucket == "" {
			bucket = cfg.Values["INFLUXDB_BUCKET"]
		}

		start, stop, err := parseTimeRange(r)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		fields, err := data.GetBooleanFieldNames()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to load boolean field names")
			return
		}
		measurement := cfg.Values["INFLUXDB_MEASUREMENT"]
		results, err := client.AggregateBooleanPercentages(measurement, bucket, fields, start, stop)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
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

		start, stop, err := parseTimeRange(r)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		// Load field lists from YAML (use cached)
		arch, err := data.GetArchitectYAML()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Server configuration error: "+err.Error())
			return
		}

		// Always use the Influx field/tag names from YAML cache for queries
		booleanFields := make([]string, 0, len(arch.BooleanFields))
		for _, f := range arch.BooleanFields {
			booleanFields = append(booleanFields, f.Name)
		}
		faultFields := make([]string, 0, len(arch.FaultFields))
		for _, f := range arch.FaultFields {
			faultFields = append(faultFields, f.Name)
		}
		// Generate the combined/namespaced float field names for InfluxDB
		floatFields := data.GetCombinedFloatFields(arch)

		measurement := cfg.Values["INFLUXDB_MEASUREMENT"]
		if measurement == "" {
			measurement = "status_data"
		}

		// Aggregate booleans (percentage true)
		boolResults, err := client.AggregateBooleanPercentages(measurement, bucket, booleanFields, start, stop)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Boolean aggregation error: "+err.Error())
			return
		}
		// Aggregate faults (count true)
		faultResults, err := client.AggregateFaultCounts(measurement, bucket, faultFields, start, stop)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Fault aggregation error: "+err.Error())
			return
		}
		// Aggregate floats (mean)
		floatResults, err := client.AggregateFloatMeans(measurement, bucket, floatFields, start, stop)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Float aggregation error: "+err.Error())
			return
		}

		results := StatsResponse{
			ProjectMeta:        arch.ProjectMeta,
			BooleanPercentages: boolResults,
			FaultCounts:        faultResults,
			FloatAverages:      floatResults,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	http.HandleFunc("/api/float-range", func(w http.ResponseWriter, r *http.Request) {
		bucket := r.URL.Query().Get("bucket")
		if bucket == "" {
			bucket = cfg.Values["INFLUXDB_BUCKET"]
		}

		field := r.URL.Query().Get("field")
		if field == "" {
			respondWithError(w, http.StatusBadRequest, "Missing required 'field' query parameter")
			return
		}

		start, stop, err := parseTimeRange(r)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Call the InfluxDB client to get the float range data
		rangeData, err := client.GetFloatRange(bucket, field, start, stop)
		if err != nil {
			log.Printf("ERROR: Error getting float range data for field '%s': %v", field, err)
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve float range data: "+err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rangeData)
	})

	http.HandleFunc("/api/upload-csv", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Limit upload size to 1MB to be safe
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			log.Println("API: Error parsing multipart form:", err)
			respondWithError(w, http.StatusBadRequest, "File is too large (max 1MB).")
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			log.Println("API: Error retrieving the file from form-data:", err)
			respondWithError(w, http.StatusBadRequest, "Error retrieving file. Make sure it's under the 'file' key.")
			return
		}
		defer file.Close()

		log.Printf("API: Received CSV upload: %s, Size: %d. Processing...", handler.Filename, handler.Size)

		// Define the destination for the final YAML file.
		yamlPath := filepath.Join(config.SharedDir, "architect.yaml")

		// Convert the uploaded CSV stream directly to YAML.
		if err := data.CSVToYAML(file, yamlPath); err != nil {
			log.Printf("API: Error converting CSV to YAML: %v", err)
			respondWithError(w, http.StatusBadRequest, "Failed to process CSV file: "+err.Error())
			return
		}

		log.Printf("API: Successfully converted CSV to %s.", yamlPath)

		// After conversion, immediately reload the configuration into the cache.
		if err := data.LoadAndCacheArchitectYAML(yamlPath); err != nil {
			log.Printf("API: CRITICAL: Converted YAML but failed to reload it: %v", err)
			// The file is updated, but the running config is stale. This is a server-side issue.
			respondWithError(w, http.StatusInternalServerError, "File converted, but server failed to apply new configuration.")
			return
		}

		log.Printf("API: New architect.yaml loaded and cached successfully.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "File '" + handler.Filename + "' uploaded, converted, and new configuration applied successfully.",
		})
	})

	// Serve the static console files
	// http.Handle("/", http.FileServer(http.Dir("../console/dist")))
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("API: Failed to create sub FS: %v", err)
	}
	http.Handle("/", http.FileServer(http.FS(subFS)))

	log.Println("API: API server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

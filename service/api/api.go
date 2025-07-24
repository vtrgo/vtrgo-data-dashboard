package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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

// respondWithError is a helper to send a JSON error message with a status code.
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func StartAPIServer(cfg *config.Config, client *influx.Client) {
	http.HandleFunc("/api/percentages", func(w http.ResponseWriter, r *http.Request) {
		bucket := r.URL.Query().Get("bucket")
		if bucket == "" {
			bucket = cfg.Values["INFLUXDB_BUCKET"]
		}

		start := r.URL.Query().Get("start")
		if start == "" {
			start = "-1h"
		} else if !isValidFluxTime(start) {
			http.Error(w, "API: Invalid start time format", http.StatusBadRequest)
			return
		}

		stop := r.URL.Query().Get("stop")
		if stop == "" {
			stop = "now()"
		} else if !isValidFluxTime(stop) {
			http.Error(w, "API: Invalid stop time format", http.StatusBadRequest)
			return
		}

		fields, err := GetBooleanFieldNames()
		if err != nil {
			http.Error(w, "API: Failed to load boolean field names", http.StatusInternalServerError)
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
			http.Error(w, "API: Invalid start time format", http.StatusBadRequest)
			return
		}

		stop := r.URL.Query().Get("stop")
		if stop == "" {
			stop = "now()"
		} else if !isValidFluxTime(stop) {
			http.Error(w, "API: Invalid stop time format", http.StatusBadRequest)
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

		// Aggregate booleans (percentage true)
		boolResults, err := client.AggregateBooleanPercentages(measurement, bucket, booleanFields, start, stop)
		if err != nil {
			http.Error(w, "API: Boolean aggregation error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Aggregate faults (count true)
		faultResults, err := client.AggregateFaultCounts(measurement, bucket, faultFields, start, stop)
		if err != nil {
			http.Error(w, "API: Fault aggregation error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Aggregate floats (mean)
		floatResults, err := client.AggregateFloatMeans(measurement, bucket, floatFields, start, stop)
		if err != nil {
			http.Error(w, "API: Float aggregation error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get project metadata from the cached YAML
		projectMeta := data.GetProjectMeta()

		results := StatsResponse{
			ProjectMeta:        projectMeta,
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
			http.Error(w, "API: Missing required 'field' query parameter", http.StatusBadRequest)
			return
		}

		start := r.URL.Query().Get("start")
		if start == "" {
			start = "-1h"
		} else if !isValidFluxTime(start) {
			http.Error(w, "API: Invalid start time format", http.StatusBadRequest)
			return
		}

		stop := r.URL.Query().Get("stop")
		if stop == "" {
			stop = "now()"
		} else if !isValidFluxTime(stop) {
			http.Error(w, "API: Invalid stop time format", http.StatusBadRequest)
			return
		}

		// Call the InfluxDB client to get the float range data
		data, err := client.GetFloatRange(bucket, field, start, stop)
		if err != nil {
			log.Printf("ERROR: Error getting float range data for field '%s': %v", field, err)
			http.Error(w, "API: Failed to retrieve float range data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	http.HandleFunc("/api/upload-csv", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Limit upload size to 10MB
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Println("API: Error parsing multipart form:", err)
			respondWithError(w, http.StatusBadRequest, "File is too large (max 10MB).")
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			log.Println("API: Error retrieving the file from form-data:", err)
			respondWithError(w, http.StatusBadRequest, "Error retrieving file. Make sure it's under the 'file' key.")
			return
		}
		defer file.Close()

		log.Printf("API: Uploading File: %s, Size: %d", handler.Filename, handler.Size)

		uploadDir := "../shared"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			log.Println("API: Error creating upload directory:", err)
			respondWithError(w, http.StatusInternalServerError, "Could not create upload directory on server.")
			return
		}

		// Sanitize filename to prevent path traversal.
		safeFilename := filepath.Base(handler.Filename)
		dstPath := filepath.Join(uploadDir, safeFilename)

		dst, err := os.Create(dstPath)
		if err != nil {
			log.Println("API: Error creating the destination file:", err)
			respondWithError(w, http.StatusInternalServerError, "Could not create file on server.")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			log.Println("API: Error copying file content:", err)
			respondWithError(w, http.StatusInternalServerError, "Could not save file content.")
			return
		}

		log.Printf("API: Successfully saved file to %s", dstPath)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "File '" + safeFilename + "' uploaded successfully.",
		})
	})

	// Serve the static console files
	http.Handle("/", http.FileServer(http.Dir("../console/dist")))

	log.Println("API: API server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

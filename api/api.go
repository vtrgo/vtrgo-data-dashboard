package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"vtarchitect/config"
	"vtarchitect/data"
	"vtarchitect/influx"
)

func collectBooleanFieldNames() []string {
	empty := data.PLCDataMap{}
	raw := influx.StructToInfluxFields(empty, "")
	fields := make([]string, 0)
	for k, v := range raw {
		if _, ok := v.(bool); ok {
			fields = append(fields, k)
		}
	}
	return fields
}

func isValidFluxTime(input string) bool {
	if input == "now()" || (len(input) > 1 && input[0] == '-') {
		return true
	}
	_, err := time.Parse(time.RFC3339, input)
	return err == nil
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

		fields := collectBooleanFieldNames()
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

		fields := collectBooleanFieldNames()
		measurement := cfg.Values["INFLUXDB_MEASUREMENT"]
		results, err := client.AggregateBooleanStats(measurement, bucket, fields, start, stop)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	log.Println("API server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

package api

import (
	"encoding/json"
	"log"
	"net/http"

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

func StartAPIServer(client *influx.Client) {
	http.HandleFunc("/api/percentages", func(w http.ResponseWriter, r *http.Request) {
		bucket := r.URL.Query().Get("bucket")
		if bucket == "" {
			bucket = "vtrFeederData"
		}
		start := r.URL.Query().Get("start")
		if start == "" {
			start = "-1h"
		}
		stop := r.URL.Query().Get("stop")
		if stop == "" {
			stop = "now()"
		}

		fields := collectBooleanFieldNames()
		log.Printf("Querying InfluxDB with bucket: %s, start: %s, stop: %s", bucket, start, stop)
		log.Printf("Boolean fields: %v", fields)
		results, err := client.AggregateBooleanPercentages(bucket, fields, start, stop)
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

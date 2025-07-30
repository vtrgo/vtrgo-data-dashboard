// file: service/influx/process.go
package influx

import (
	"log"
	"time"
	"vtarchitect/config"
	"vtarchitect/utils"
)

// ProcessAndLogChangedData writes only changed fields to InfluxDB using the YAML-driven map, recursively.
func ProcessAndLogChangedData(cfg *config.Config, plcData, prev map[string]interface{}, batchWriter *ChannelBatchWriter) {
	measurement := cfg.Values["INFLUXDB_MEASUREMENT"]
	if measurement == "" {
		measurement = "status_data"
	}
	changed := make(map[string]interface{})
	utils.CollectChangedFields(plcData, prev, changed, "")
	if len(changed) == 0 {
		return // nothing to write
	}
	batchWriter.AddPoint(measurement, nil, changed, time.Now())
	log.Printf("INFLUX: Buffered changed fields for InfluxDB: %s", changed)
}

// ProcessAndLogFullData writes the full PLC state to InfluxDB using the YAML-driven map.
func ProcessAndLogFullData(cfg *config.Config, plcData map[string]interface{}, batchWriter *ChannelBatchWriter) {
	measurement := cfg.Values["INFLUXDB_MEASUREMENT"]
	if measurement == "" {
		measurement = "status_data"
	}
	batchWriter.AddPoint(measurement, nil, plcData, time.Now())
	log.Println("INFLUX: Buffered full-state write for InfluxDB")
}

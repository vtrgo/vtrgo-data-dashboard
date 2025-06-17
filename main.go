// File: main.go
// Package main implements a Modbus TCP server that collects boolean field data
// and writes it to InfluxDB only if the data has changed.
package main

import (
	"log"
	"strconv"
	"time"

	"vtarchitect/api"
	"vtarchitect/config"
	"vtarchitect/data"
	"vtarchitect/influx"

	"github.com/tbrandon/mbserver"
)

// collectBooleanFieldNames retrieves the names of boolean fields from an empty PLC data map.
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

// hasChanges checks if there are any boolean field changes between previous and current PLC data.
func hasChanges(prev, curr data.PLCDataMap) bool {
	p := influx.StructToInfluxFields(prev, "")
	c := influx.StructToInfluxFields(curr, "")
	for k, v := range c {
		if pv, ok := p[k]; ok {
			if vb, ok := v.(bool); ok {
				if pb, ok := pv.(bool); ok && vb != pb {
					return true
				}
			}
		}
	}
	return false
}

// changedFields returns a map of only the boolean fields that have changed between prev and curr.
func changedFields(prev, curr data.PLCDataMap) map[string]interface{} {
	p := influx.StructToInfluxFields(prev, "")
	c := influx.StructToInfluxFields(curr, "")
	changed := make(map[string]interface{})
	for k, v := range c {
		if pv, ok := p[k]; ok {
			if vb, ok := v.(bool); ok {
				if pb, ok := pv.(bool); ok && vb != pb {
					changed[k] = vb
				}
			}
		}
	}
	return changed
}

// processAndLogChanged writes only changed boolean fields to InfluxDB.
func processAndLogChanged(cfg *config.Config, plcData data.PLCDataMap, influxClient *influx.Client, prev data.PLCDataMap) {
	measurement := cfg.Values["INFLUXDB_MEASUREMENT"]
	if measurement == "" {
		measurement = "status_data"
	}
	fields := changedFields(prev, plcData)
	if len(fields) == 0 {
		return // nothing to write
	}
	err := influxClient.WritePoint(measurement, nil, fields, time.Now())
	log.Printf("Writing changed fields to InfluxDB: %s", fields)
	if err != nil {
		log.Printf("Error writing to InfluxDB: %v", err)
	}
}

// getPollInterval retrieves the polling interval from the configuration.
func getPollInterval(cfg *config.Config) time.Duration {
	pollInterval := cfg.Values["PLC_POLL_MS"]
	pollIntervalMs, err := strconv.Atoi(pollInterval)
	if err != nil || pollIntervalMs <= 0 {
		pollIntervalMs = 1000 // default to 1 second
	}
	return time.Duration(pollIntervalMs) * time.Millisecond
}

// getFullWriteInterval retrieves the full-state write interval from the configuration (in minutes).
func getFullWriteInterval(cfg *config.Config) time.Duration {
	intervalStr := cfg.Values["FULL_WRITE_MINUTES"]
	intervalMin, err := strconv.Atoi(intervalStr)
	if err != nil || intervalMin <= 0 {
		intervalMin = 60 // default to 60 minutes
	}
	return time.Duration(intervalMin) * time.Minute
}

// processAndLogFull writes the full PLC state to InfluxDB, regardless of changes.
func processAndLogFull(cfg *config.Config, plcData data.PLCDataMap, influxClient *influx.Client) {
	measurement := cfg.Values["INFLUXDB_MEASUREMENT"]
	if measurement == "" {
		measurement = "status_data"
	}
	fields := influx.StructToInfluxFields(plcData, "")
	log.Println("DEBUG: Full-state InfluxDB write fields:")
	for k, v := range fields {
		log.Printf("  %s: %v", k, v)
	}
	err := influxClient.WritePoint(measurement, nil, fields, time.Now())
	log.Printf("Full-state write to InfluxDB: %s", fields)
	if err != nil {
		log.Printf("Error writing full state to InfluxDB: %v", err)
	}
}

// runEthernetIPCycle connects to the PLC via Ethernet/IP and continuously polls for data changes.
func runEthernetIPCycle(cfg *config.Config, influxClient *influx.Client) {
	ip := cfg.Values["ETHERNET_IP_ADDRESS"]
	eth := data.NewPLC(ip)

	for {
		err := eth.Connect()
		if err != nil {
			log.Printf("PLC connection failed, retrying in 5 seconds: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	defer eth.Disconnect()

	pollInterval := getPollInterval(cfg)
	fullWriteInterval := getFullWriteInterval(cfg)
	fullWriteTicker := time.NewTicker(fullWriteInterval)
	defer fullWriteTicker.Stop()
	var last data.PLCDataMap
	for {
		plcData := data.LoadFromEthernetIP(cfg, eth)
		select {
		case <-fullWriteTicker.C:
			processAndLogFull(cfg, plcData, influxClient)
			last = plcData
		default:
			if hasChanges(last, plcData) {
				processAndLogChanged(cfg, plcData, influxClient, last)
				last = plcData
			}
		}
		time.Sleep(pollInterval)
	}
}

// runModbusCycle connects to the Modbus TCP server and continuously polls for data changes.
func runModbusCycle(cfg *config.Config, server *mbserver.Server, influxClient *influx.Client) {
	startStr := cfg.Values["MODBUS_REGISTER_START"]
	endStr := cfg.Values["MODBUS_REGISTER_END"]
	start, err := strconv.Atoi(startStr)
	if err != nil {
		log.Fatalf("Invalid MODBUS_REGISTER_START: %v", err)
	}
	end, err := strconv.Atoi(endStr)
	if err != nil {
		log.Fatalf("Invalid MODBUS_REGISTER_END: %v", err)
	}

	pollInterval := getPollInterval(cfg)
	fullWriteInterval := getFullWriteInterval(cfg)
	fullWriteTicker := time.NewTicker(fullWriteInterval)
	defer fullWriteTicker.Stop()
	var last data.PLCDataMap
	for {
		if len(server.HoldingRegisters) <= end {
			log.Println("Insufficient register length, skipping cycle")
			time.Sleep(5 * time.Second)
			continue
		}
		readSlice := server.HoldingRegisters[start : end+1]
		plcData := data.LoadPLCDataMap(cfg, readSlice)
		select {
		case <-fullWriteTicker.C:
			processAndLogFull(cfg, plcData, influxClient)
			last = plcData
		default:
			if hasChanges(last, plcData) {
				processAndLogChanged(cfg, plcData, influxClient, last)
				last = plcData
			}
		}
		time.Sleep(pollInterval)
	}
}

// main initializes the application, loads configuration, sets up InfluxDB client,
// starts the API server, and runs the appropriate PLC data collection cycle.
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	plcSource := cfg.Values["PLC_DATA_SOURCE"]
	log.Printf("PLC data source: %s", plcSource)

	influxClient, err := influx.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to InfluxDB: %v", err)
	}
	defer influxClient.Close()

	go api.StartAPIServer(cfg, influxClient)

	boolFields := collectBooleanFieldNames()
	// fmt.Println("Boolean fields for aggregation:")
	for _, field := range boolFields {
		log.Println(field)
	}

	if plcSource == "ethernet-ip" {
		runEthernetIPCycle(cfg, influxClient)
	} else {
		server := mbserver.NewServer()
		port := cfg.Values["MODBUS_TCP_PORT"]
		if port == "" {
			port = "5020"
		}
		err := server.ListenTCP("0.0.0.0:" + port)
		if err != nil {
			log.Fatalf("Failed to start Modbus server: %v", err)
		}
		defer server.Close()
		log.Printf("Modbus server listening on port %s", port)

		runModbusCycle(cfg, server, influxClient)
	}
}

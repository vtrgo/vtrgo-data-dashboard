// File: main.go
// Package main implements a Modbus TCP server that collects boolean field data
// and writes it to InfluxDB.
// It also queries InfluxDB for boolean field percentages over the last minute
// and logs the results.
package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"vtarchitect/api"
	"vtarchitect/config"
	"vtarchitect/data"
	"vtarchitect/influx"

	"github.com/tbrandon/mbserver"
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

func processAndLog(cfg *config.Config, plcData data.PLCDataMap, influxClient *influx.Client, boolFields []string) {
	measurement := cfg.Values["INFLUXDB_MEASUREMENT"]
	if measurement == "" {
		measurement = "status_data"
	}
	fields := influx.StructToInfluxFields(plcData, "")
	err := influxClient.WritePoint(measurement, nil, fields, time.Now())
	if err != nil {
		log.Printf("Error writing to InfluxDB: %v", err)
	}

	// percentages, err := influxClient.AggregateBooleanPercentages(bucket, boolFields, "-1m", "now()")
	// if err != nil {
	// 	log.Printf("Error querying InfluxDB: %v", err)
	// } else {
	// 	// log.Println("Boolean field true percentages over the last 1 minute:")
	// 	for field, pct := range percentages {
	// 		log.Printf("%s: %.2f%%", field, pct)
	// 	}
	// }
}

func getPollInterval(cfg *config.Config) time.Duration {
	pollMsStr := cfg.Values["PLC_POLL_MS"]
	pollMs, err := strconv.Atoi(pollMsStr)
	if err != nil || pollMs <= 0 {
		pollMs = 1000 // default to 1 second
	}
	return time.Duration(pollMs) * time.Millisecond
}

func runEthernetIPCycle(cfg *config.Config, influxClient *influx.Client, boolFields []string) {
	ip := cfg.Values["PLC_ETHERNET_IP_ADDRESS"]
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
	for {
		plcData := data.LoadFromEthernetIP(eth)
		processAndLog(cfg, plcData, influxClient, boolFields)
		time.Sleep(pollInterval)
	}
}

func runModbusCycle(cfg *config.Config, server *mbserver.Server, influxClient *influx.Client, boolFields []string) {
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
	for {
		if len(server.HoldingRegisters) <= end {
			log.Println("Insufficient register length, skipping cycle")
			time.Sleep(5 * time.Second)
			continue
		}
		readSlice := server.HoldingRegisters[start : end+1]
		plcData := data.LoadPLCDataMap(readSlice)
		processAndLog(cfg, plcData, influxClient, boolFields)
		time.Sleep(pollInterval)
	}
}

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
	fmt.Println("Boolean fields for aggregation:")
	for _, field := range boolFields {
		log.Println(field)
	}

	if plcSource == "ethernet-ip" {
		runEthernetIPCycle(cfg, influxClient, boolFields)
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

		runModbusCycle(cfg, server, influxClient, boolFields)
	}
}

// File: main.go
// Package main implements a Modbus TCP server that collects boolean field data
// and writes it to InfluxDB.
// It also queries InfluxDB for boolean field percentages over the last minute
// and logs the results.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"vtarchitect/api"
	"vtarchitect/data"
	"vtarchitect/influx"

	"github.com/joho/godotenv"
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

func processAndLog(plcData data.PLCDataMap, influxClient *influx.Client, boolFields []string) {
	measurement := os.Getenv("INFLUXDB_MEASUREMENT")
	// bucket := os.Getenv("INFLUXDB_BUCKET")

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

func getPollInterval() time.Duration {
	pollMsStr := os.Getenv("PLC_POLL_MS")
	pollMs, err := strconv.Atoi(pollMsStr)
	if err != nil || pollMs <= 0 {
		pollMs = 1000 // default to 1 second
	}
	return time.Duration(pollMs) * time.Millisecond
}

func runEthernetIPCycle(influxClient *influx.Client, boolFields []string) {
	ip := os.Getenv("PLC_ETHERNET_IP_ADDRESS")
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

	pollInterval := getPollInterval()
	for {
		plcData := data.LoadFromEthernetIP(eth)
		processAndLog(plcData, influxClient, boolFields)
		time.Sleep(pollInterval)
	}
}

func runModbusCycle(server *mbserver.Server, influxClient *influx.Client, boolFields []string) {
	startStr := os.Getenv("MODBUS_REGISTER_START")
	endStr := os.Getenv("MODBUS_REGISTER_END")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		log.Fatalf("Invalid MODBUS_REGISTER_START: %v", err)
	}
	end, err := strconv.Atoi(endStr)
	if err != nil {
		log.Fatalf("Invalid MODBUS_REGISTER_END: %v", err)
	}

	pollInterval := getPollInterval()
	for {
		if len(server.HoldingRegisters) <= end {
			log.Println("Insufficient register length, skipping cycle")
			time.Sleep(5 * time.Second)
			continue
		}
		readSlice := server.HoldingRegisters[start : end+1]
		plcData := data.LoadPLCDataMap(readSlice)
		processAndLog(plcData, influxClient, boolFields)
		time.Sleep(pollInterval)
	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env")
	}

	plcSource := os.Getenv("PLC_DATA_SOURCE")
	log.Printf("PLC data source: %s", plcSource)

	influxClient, err := influx.NewClient()
	if err != nil {
		log.Fatalf("Failed to connect to InfluxDB: %v", err)
	}
	defer influxClient.Close()

	go api.StartAPIServer(influxClient)

	boolFields := collectBooleanFieldNames()
	fmt.Println("Boolean fields for aggregation:")
	for _, field := range boolFields {
		log.Println(field)
	}

	if plcSource == "ethernet-ip" {
		runEthernetIPCycle(influxClient, boolFields)
	} else {
		server := mbserver.NewServer()
		port := os.Getenv("MODBUS_TCP_PORT")
		if port == "" {
			port = "5020"
		}
		err := server.ListenTCP("0.0.0.0:" + port)
		if err != nil {
			log.Fatalf("Failed to start Modbus server: %v", err)
		}
		defer server.Close()
		log.Printf("Modbus server listening on port %s", port)

		runModbusCycle(server, influxClient, boolFields)
	}
}

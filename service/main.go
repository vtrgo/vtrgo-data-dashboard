// file: service/main.go
package main

import (
	"log"
	"path/filepath"

	"vtarchitect/api"
	"vtarchitect/config"
	"vtarchitect/data"
	"vtarchitect/influx"

	"github.com/tbrandon/mbserver"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: Failed to load config: %v", err)
	}

	log.Println("STARTUP: Loading and caching architect.yaml...")
	architectPath := filepath.Join(config.SharedDir, "architect.yaml")
	err = data.LoadAndCacheArchitectYAML(architectPath)
	if err != nil {
		log.Fatalf("FATAL: Failed to load architect.yaml: %v", err)
	}
	log.Println("STARTUP: architect.yaml loaded and cached successfully.")

	plcSource := cfg.Values["PLC_DATA_SOURCE"]
	log.Printf("STARTUP: PLC data source: %s", plcSource)

	influxClient, err := influx.NewClient(cfg)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to InfluxDB: %v", err)
	}
	defer influxClient.Close()

	batchWriter := influx.NewChannelBatchWriter(influxClient.GetWriteAPI(), 100)
	defer batchWriter.Close()

	go api.StartAPIServer(cfg, influxClient)

	if plcSource == "ethernet-ip" {
		data.RunEthernetIPCycle(cfg, batchWriter)
	} else {
		server := mbserver.NewServer()
		port := cfg.Values["MODBUS_TCP_PORT"]
		if port == "" {
			port = "5020"
		}
		err := server.ListenTCP("0.0.0.0:" + port)
		if err != nil {
			log.Fatalf("FATAL: Failed to start Modbus server: %v", err)
		}
		defer server.Close()
		log.Printf("DATA: Modbus server listening on port %s", port)

		data.RunModbusCycle(cfg, server, batchWriter)
	}
}

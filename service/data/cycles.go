// service/data/cycles.go
// Cycles for continuously polling PLC data and writing to InfluxDB
package data

import (
	"log"
	"strconv"
	"time"
	"vtarchitect/config"
	"vtarchitect/influx"
	"vtarchitect/utils"

	"github.com/tbrandon/mbserver"
)

// runEthernetIPCycle connects to the PLC via Ethernet/IP and continuously polls for data changes.
func RunEthernetIPCycle(cfg *config.Config, batchWriter *influx.ChannelBatchWriter) {
	ip := cfg.Values["ETHERNET_IP_ADDRESS"]
	eth := NewPLC(ip)

	for {
		err := eth.Connect()
		if err != nil {
			log.Printf("DATA: PLC connection failed, retrying in 5 seconds: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	defer eth.Disconnect()

	pollInterval := utils.GetPollInterval(cfg)
	fullWriteInterval := utils.GetFullWriteInterval(cfg)
	fullWriteTicker := time.NewTicker(fullWriteInterval)
	defer fullWriteTicker.Stop()
	var last map[string]interface{}
	for {
		plcData, err := LoadFromEthernetIP(cfg, eth)
		if err != nil {
			log.Printf("DATA:Error loading PLC data from Ethernet/IP YAML: %v", err)
			time.Sleep(pollInterval)
			continue
		}
		select {
		case <-fullWriteTicker.C:
			influx.ProcessAndLogFullData(cfg, plcData, batchWriter)
			last = plcData
		default:
			if !utils.MapsEqual(last, plcData) {
				influx.ProcessAndLogChangedData(cfg, plcData, last, batchWriter)
				last = plcData
			}
		}
		time.Sleep(pollInterval)
	}
}

// runModbusCycle connects to the Modbus TCP server and continuously polls for data changes.
func RunModbusCycle(cfg *config.Config, server *mbserver.Server, batchWriter *influx.ChannelBatchWriter) {
	startStr := cfg.Values["MODBUS_REGISTER_START"]
	endStr := cfg.Values["MODBUS_REGISTER_END"]
	start, err := strconv.Atoi(startStr)
	if err != nil {
		log.Fatalf("FATAL: Invalid MODBUS_REGISTER_START: %v", err)
	}
	end, err := strconv.Atoi(endStr)
	if err != nil {
		log.Fatalf("FATAL: Invalid MODBUS_REGISTER_END: %v", err)
	}

	pollInterval := utils.GetPollInterval(cfg)
	fullWriteInterval := utils.GetFullWriteInterval(cfg)
	fullWriteTicker := time.NewTicker(fullWriteInterval)
	defer fullWriteTicker.Stop()
	var last map[string]interface{}
	for {
		if len(server.HoldingRegisters) <= end {
			log.Println("DATA: Insufficient register length, skipping cycle")
			time.Sleep(5 * time.Second)
			continue
		}
		readSlice := server.HoldingRegisters[start : end+1]
		plcData, err := ParsePLCDataFromRegisters(readSlice)
		if err != nil {
			log.Printf("ERROR: Error loading PLC data from Modbus YAML: %v", err)
			time.Sleep(pollInterval)
			continue
		}
		select {
		case <-fullWriteTicker.C:
			influx.ProcessAndLogFullData(cfg, plcData, batchWriter)
			last = plcData
		default:
			if !utils.MapsEqual(last, plcData) {
				influx.ProcessAndLogChangedData(cfg, plcData, last, batchWriter)
				last = plcData
			}
		}
		time.Sleep(pollInterval)
	}
}

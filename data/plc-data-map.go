package data

import (
	"fmt"
	"math"
	"os"
	"strings"
	"vtarchitect/config"

	yaml "gopkg.in/yaml.v3"
)

// PLCDataMap holds the structure of the PLC data map, including system status bits, feeder status bits, level status bits, jam status bits, fault status bits, and other relevant data.
type PLCDataMap struct {
	SystemStatusBits struct {
		ControlPowerON bool
		AutoMode       bool
		PurgeMode      bool
		SystemIdle     bool
		AirPressureOK  bool
		SystemFaulted  bool
	}
	FeederStatusBits struct {
		BulkHopperEnabled       bool
		BulkHopperLevelNotOK    bool
		ElevatorEnabled         bool
		ElevatorLevelNotOK      bool
		CrossConveyorEnabled    bool
		CrossConveyorLevelNotOK bool
		OrientationEnabled      bool
		OrientationLevelNotOK   bool
		TransferEnabled         bool
		TransferLevelNotOK      bool
		EscapementAdvEnabled    bool
		EscapementRetEnabled    bool
	}
	LevelStatusBits struct {
		HighLevelLane   [8]bool `influx:"Lane"`
		NotLowLevelLane [8]bool `influx:"Lane"`
	}
	JamStatusBits struct {
		JamInOrientation [8]bool `influx:"Lane"`
	}
	FaultStatusBits struct {
		FaultArray0 [16]bool `influx:"Fault"`
		FaultArray1 [16]bool `influx:"Fault"`
	}

	VibrationDataFloats [5]VibrationData // mapped from OtherStatusWords
}

// VibrationData represents the structure of vibration data read from the PLC.
type VibrationData struct {
	VibrationX  float32
	VibrationY  float32
	VibrationZ  float32
	Temperature float32
}

// LoadPLCDataMap reads PLC data from the provided registers and returns a PLCDataMap.
func LoadPLCDataMap(cfg *config.Config, registers []uint16) PLCDataMap {
	var m PLCDataMap
	lengthStr, ok := cfg.Values["ETHERNET_IP_LENGTH"]
	if !ok {
		return m // or handle error/log as needed
	}
	var length int
	fmt.Sscanf(lengthStr, "%d", &length)
	if len(registers) < length {
		return m // or handle error/log as needed
	}

	// 1000: SystemStatusBits
	word := registers[0]
	m.SystemStatusBits.ControlPowerON = word&(1<<0) != 0
	m.SystemStatusBits.AutoMode = word&(1<<1) != 0
	m.SystemStatusBits.PurgeMode = word&(1<<2) != 0
	m.SystemStatusBits.SystemIdle = word&(1<<3) != 0
	m.SystemStatusBits.AirPressureOK = word&(1<<4) != 0
	m.SystemStatusBits.SystemFaulted = word&(1<<5) != 0

	// 1001: FeederStatusBits
	word = registers[1]
	m.FeederStatusBits.BulkHopperEnabled = word&(1<<0) != 0
	m.FeederStatusBits.BulkHopperLevelNotOK = word&(1<<1) != 0
	m.FeederStatusBits.ElevatorEnabled = word&(1<<2) != 0
	m.FeederStatusBits.ElevatorLevelNotOK = word&(1<<3) != 0
	m.FeederStatusBits.CrossConveyorEnabled = word&(1<<4) != 0
	m.FeederStatusBits.CrossConveyorLevelNotOK = word&(1<<5) != 0
	m.FeederStatusBits.OrientationEnabled = word&(1<<6) != 0
	m.FeederStatusBits.OrientationLevelNotOK = word&(1<<7) != 0
	m.FeederStatusBits.TransferEnabled = word&(1<<8) != 0
	m.FeederStatusBits.TransferLevelNotOK = word&(1<<9) != 0
	m.FeederStatusBits.EscapementAdvEnabled = word&(1<<10) != 0
	m.FeederStatusBits.EscapementRetEnabled = word&(1<<11) != 0

	// 1002: LevelStatusBits
	word = registers[2]
	for i := 0; i < 8; i++ {
		m.LevelStatusBits.HighLevelLane[i] = word&(1<<i) != 0
		m.LevelStatusBits.NotLowLevelLane[i] = word&(1<<(i+8)) != 0
	}

	// 1003: JamStatusBits
	word = registers[3]
	for i := 0; i < 8; i++ {
		m.JamStatusBits.JamInOrientation[i] = word&(1<<i) != 0
	}

	// 1004: FaultBits
	word = registers[4]
	for i := 0; i < 16; i++ {
		m.FaultStatusBits.FaultArray0[i] = word&(1<<i) != 0
	}
	word = registers[5]
	for i := 0; i < 16; i++ {
		m.FaultStatusBits.FaultArray1[i] = word&(1<<i) != 0
	}

	// 1070–1109: VibrationDataWords
	for i := 0; i < 5; i++ {
		baseIdx := 70 + i*8                                                                                                          // Each group of 4 float32 values occupies 8 registers
		m.VibrationDataFloats[i].VibrationX = math.Float32frombits(uint32(registers[baseIdx])<<16 | uint32(registers[baseIdx+1]))    // VibrationX
		m.VibrationDataFloats[i].VibrationY = math.Float32frombits(uint32(registers[baseIdx+2])<<16 | uint32(registers[baseIdx+3]))  // VibrationY
		m.VibrationDataFloats[i].VibrationZ = math.Float32frombits(uint32(registers[baseIdx+4])<<16 | uint32(registers[baseIdx+5]))  // VibrationZ
		m.VibrationDataFloats[i].Temperature = math.Float32frombits(uint32(registers[baseIdx+6])<<16 | uint32(registers[baseIdx+7])) // Temperature
	}
	return m
}

// ArchitectYAML represents the structure of architect.yaml
type ArchitectYAML struct {
	BooleanFields []struct {
		Name    string `yaml:"name"`
		Address int    `yaml:"address"`
		Bit     *int   `yaml:"bit,omitempty"`
	} `yaml:"boolean_fields"`
	FaultFields []struct {
		Name    string `yaml:"name"`
		Address int    `yaml:"address"`
		Bit     *int   `yaml:"bit,omitempty"`
	} `yaml:"fault_fields"`
	FloatFields []struct {
		Name    string `yaml:"name"`
		Address int    `yaml:"address"`
	} `yaml:"float_fields"`
}

// LoadPLCDataMapFromYAML loads the PLCDataMap from registers using architect.yaml mapping (generic for booleans, faults, floats)
func LoadPLCDataMapFromYAML(yamlPath string, registers []uint16) (map[string]interface{}, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}
	var arch ArchitectYAML
	err = yaml.Unmarshal(data, &arch)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	// Booleans
	for _, field := range arch.BooleanFields {
		reg := registers[field.Address]
		bit := 0
		if field.Bit != nil {
			bit = *field.Bit
		}
		val := (reg & (1 << bit)) != 0
		result[field.Name] = val
	}

	// Faults
	for _, field := range arch.FaultFields {
		reg := registers[field.Address]
		bit := 0
		if field.Bit != nil {
			bit = *field.Bit
		}
		val := (reg & (1 << bit)) != 0
		result[field.Name] = val
	}

	// Floats (robust pairing of HighINT/LowINT)
	floatPairs := make(map[string]struct{ High, Low *int })
	for i := 0; i < len(arch.FloatFields); i++ {
		name := arch.FloatFields[i].Name
		addr := arch.FloatFields[i].Address
		if strings.HasSuffix(name, "(HighINT)") {
			base := strings.TrimSuffix(name, "(HighINT)")
			p := floatPairs[base]
			p.High = &addr
			floatPairs[base] = p
		} else if strings.HasSuffix(name, "(LowINT)") {
			base := strings.TrimSuffix(name, "(LowINT)")
			p := floatPairs[base]
			p.Low = &addr
			floatPairs[base] = p
		}
	}
	for base, p := range floatPairs {
		if p.High != nil && p.Low != nil {
			high := uint32(registers[*p.High])
			low := uint32(registers[*p.Low])
			f := math.Float32frombits((high << 16) | low)
			result[base] = f
		}
	}

	return result, nil
}

// LoadArchitectYAML loads and parses architect.yaml and returns the parsed ArchitectYAML struct.
func LoadArchitectYAML() (*ArchitectYAML, error) {
	data, err := os.ReadFile("data/architect.yaml")
	if err != nil {
		return nil, err
	}
	var arch ArchitectYAML
	err = yaml.Unmarshal(data, &arch)
	if err != nil {
		return nil, err
	}
	return &arch, nil
}

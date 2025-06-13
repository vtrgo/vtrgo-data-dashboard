package data

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
	SystemStatusWords struct {
		TimeInAutoMinutes   uint16
		TimeInAutoSeconds   uint16
		TimeFaultedMinutes  uint16
		TimeFaultedSeconds  uint16
		FaultCountAny       uint16
		LastCycleTimeMS     uint16
		AverageCycleTimeMS  uint16
		BinEmptyTimeMinutes uint16
		AirTrackBlowerSpeed uint16
	}
	LowLevelTimes    [8]uint16
	FaultCounts      [32]uint16
	OtherStatusWords [30]uint16 // 1070–1099
}

func LoadPLCDataMap(registers []uint16) PLCDataMap {
	var m PLCDataMap

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

	// 1010–1018: SystemStatusWords
	m.SystemStatusWords.TimeInAutoMinutes = registers[10]
	m.SystemStatusWords.TimeInAutoSeconds = registers[11]
	m.SystemStatusWords.TimeFaultedMinutes = registers[12]
	m.SystemStatusWords.TimeFaultedSeconds = registers[13]
	m.SystemStatusWords.FaultCountAny = registers[14]
	m.SystemStatusWords.LastCycleTimeMS = registers[15]
	m.SystemStatusWords.AverageCycleTimeMS = registers[16]
	m.SystemStatusWords.BinEmptyTimeMinutes = registers[17]
	m.SystemStatusWords.AirTrackBlowerSpeed = registers[18]

	// 1020–1027: LowLevelTimes
	copy(m.LowLevelTimes[:], registers[20:28])

	// 1030–1061: FaultCounts
	copy(m.FaultCounts[:], registers[30:62])

	// 1070–1099: OtherStatusWords
	copy(m.OtherStatusWords[:], registers[70:100])

	return m
}

package safety

import "fmt"

type VoltageStatus int

const (
	VoltageOK      VoltageStatus = iota
	VoltageWarning
	VoltageBlocked
)

type OperationType int

const (
	OpCoding      OperationType = iota
	OpMultiCoding
	OpFlash
)

type VoltageResult struct {
	Status  VoltageStatus
	Voltage float64
	Message string
}

type thresholds struct {
	block   float64
	warning float64
}

var voltageThresholds = map[OperationType]thresholds{
	OpCoding:      {block: 12.0, warning: 12.5},
	OpMultiCoding: {block: 12.5, warning: 12.5},
	OpFlash:       {block: 13.0, warning: 13.0},
}

func CheckVoltage(voltage float64, op OperationType) VoltageResult {
	t := voltageThresholds[op]

	if voltage < t.block {
		return VoltageResult{
			Status:  VoltageBlocked,
			Voltage: voltage,
			Message: fmt.Sprintf("Battery voltage %.1fV is below minimum %.1fV — operation blocked", voltage, t.block),
		}
	}

	if voltage > t.block && voltage < t.warning {
		return VoltageResult{
			Status:  VoltageWarning,
			Voltage: voltage,
			Message: fmt.Sprintf("Battery voltage %.1fV is low (recommended: %.1fV+) — proceed with caution", voltage, t.warning),
		}
	}

	return VoltageResult{
		Status:  VoltageOK,
		Voltage: voltage,
		Message: fmt.Sprintf("Battery voltage %.1fV OK", voltage),
	}
}
